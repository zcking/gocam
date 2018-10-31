package main

import (
	"encoding/json"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hybridgroup/mjpeg"
	"github.com/spf13/viper"
	"gocv.io/x/gocv"
)

const tempStoragePrefix string = "TMP_"

var (
	deviceID   int
	err        error
	webcam     *gocv.VideoCapture
	stream     *mjpeg.Stream
	xmlFile    string
	classifier gocv.CascadeClassifier
	blue       color.RGBA
)

var (
	img gocv.Mat
	mut sync.Mutex
)

var (
	isRunning = true
	runMut    sync.Mutex
)

type PowerResponse struct {
	PowerOn bool
}

var writeMut sync.Mutex

func main() {
	defer func() { log.Println("Gocam shutting down...") }()

	// Parse the configuration file
	viper.SetConfigName("default")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("[ERROR]: Failed to configure GoCam: %s\n", err)
		os.Exit(1)
	}

	// Set defaults for configuration
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("port", 5000)
	viper.SetDefault("captureDevice", 0)
	viper.SetDefault("facialDetectionFile", filepath.Join("data", ""))
	viper.SetDefault("tempRecLength", "0m")
	viper.SetDefault("tempKeepTime", "0m")
	viper.SetDefault("contrast", 0.5)
	viper.SetDefault("saturation", 0.75)
	viper.SetDefault("fps", 20)
	viper.SetDefault("brightness", 0.6)

	// Parse arguments
	deviceID = viper.GetInt("captureDevice")
	xmlFile = viper.GetString("facialDetectionFile")
	host := viper.GetString("host") + ":" + viper.GetString("port")
	tempRecLength, _ := time.ParseDuration(viper.GetString("tempRecLength"))
	tempKeepTime, _ := time.ParseDuration(viper.GetString("tempKeepTime"))

	// Color for the rect when faces detected
	blue = color.RGBA{B: 255}

	// Open webcam
	webcam, err = gocv.VideoCaptureDevice(int(deviceID))
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer webcam.Close()

	// Setup OS signal trapping to do proper cleanup of webcam on exit
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	signal.Stop(sigCh)
	go func() {
		for sig := range sigCh {
			// sig is a ^C, handle it
			log.Fatalf("Received Signal: %v\n", sig)
			log.Println("Waiting for 2 seconds to finish shutting down...")
			log.Println("Gocam shutting down...")
			webcam.Close()
			time.Sleep(2 * time.Second)
			os.Exit(0)
		}
	}()

	// Video capture settings
	log.Printf("Video capture configured with codec %q\n", webcam.CodecString())
	webcam.Set(gocv.VideoCaptureSaturation, viper.GetFloat64("saturation"))
	webcam.Set(gocv.VideoCaptureFPS, viper.GetFloat64("fps"))
	webcam.Set(gocv.VideoCaptureBrightness, viper.GetFloat64("brightness"))
	webcam.Set(gocv.VideoCaptureContrast, viper.GetFloat64("contrast"))

	// Prepare image matrix
	img = gocv.NewMat()
	defer img.Close()

	// Create the mjpeg stream
	stream = mjpeg.NewStream()

	// Enable face detection
	// Load classifier to recognize faces
	if xmlFile != "" {
		classifier = gocv.NewCascadeClassifier()
		defer classifier.Close()

		if !classifier.Load(xmlFile) {
			log.Fatalf("Error reading cascade file: %v\n", xmlFile)
			return
		}
	} else {
		log.Println("[WARN]: No facial detection data file provided; facial detection disabled.")
	}

	// Capture a single image just to initialize the image variable
	captureImage()

	// Capture images from the camera in parallel
	go func() {
		for {
			runMut.Lock()
			r := isRunning
			runMut.Unlock()
			if r {
				captureImage()
				if xmlFile != "" {
					detectFaces()
				}
			}
		}
	}()

	// Start capturing for mjpeg stream
	go func() {
		for {
			runMut.Lock()
			r := isRunning
			runMut.Unlock()
			if r {
				mjpegCapture()
			}
		}
	}()

	// Output temporary files to local file system
	if tempRecLength > 0 {
		go func() {
			for {
				writeMut.Lock()
				writeTemporaryStorage(tempRecLength)
				writeMut.Unlock()
			}
		}()
	} else {
		log.Println("[WARN]: temp recording length set to 0; recording will not be saved to file system.")
	}

	// Purge any older temporary files (beyond the keep time)
	if tempKeepTime > 0 {
		go purgeTemporaryStorage(tempKeepTime)
	} else {
		log.Println("[WARN]: temp keep time set to 0; any recordings saved to file system will not be erased.")
	}

	// Spin up the controller server
	http.HandleFunc("/health", func(w http.ResponseWriter, request *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/api/power/off", func(w http.ResponseWriter, request *http.Request) {
		setupResponse(&w, request)
		runMut.Lock()

		// Shutdown the camera
		isRunning = false
		runMut.Unlock()
		log.Println("Gocam powering off...")

		data := PowerResponse{isRunning}
		js, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
		}
	})

	http.HandleFunc("/api/power/on", func(w http.ResponseWriter, request *http.Request) {
		setupResponse(&w, request)
		runMut.Lock()

		// Poweron the camera
		isRunning = true
		runMut.Unlock()
		log.Printf("Gocam powering on... Streaming to %v\n", host)

		data := PowerResponse{isRunning}
		js, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
		}
	})

	http.HandleFunc("/api/power", func(w http.ResponseWriter, request *http.Request) {
		setupResponse(&w, request)
		data := PowerResponse{isRunning}
		js, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
		}
	})

	http.HandleFunc("/cam", func(w http.ResponseWriter, request *http.Request) {
		runMut.Lock()
		r := isRunning
		runMut.Unlock()
		if !r {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Camera is powered off currently."))
		} else {
			stream.ServeHTTP(w, request)
		}
	})
	log.Fatal(http.ListenAndServe(host, nil))
}

func detectFaces() {
	mut.Lock()
	defer mut.Unlock()

	// Detect faces
	rects := classifier.DetectMultiScale(img)
	//log.Printf("found %d faces\n", len(rects))

	// Draw a rectangle around each face on the original image,
	// along with text identifying as "Human"
	for _, r := range rects {
		gocv.Rectangle(&img, r, blue, 3)
		size := gocv.GetTextSize("Human", gocv.FontHersheyPlain, 1.2, 2)
		pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
		gocv.PutText(&img, "Human", pt, gocv.FontHersheyPlain, 1.2, blue, 2)
	}
}

func captureImage() {
	mut.Lock()
	defer mut.Unlock()

	if ok := webcam.Read(&img); !ok {
		log.Fatalf("Device closed: %v\n", deviceID)
		syscall.Exit(-1)
	}
	if img.Empty() {
		syscall.Exit(-1)
	}
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func mjpegCapture() {
	mut.Lock()
	buf, _ := gocv.IMEncode(".jpg", img)
	stream.UpdateJPEG(buf)
	mut.Unlock()
}

func writeTemporaryStorage(interval time.Duration) {
	runMut.Lock()
	r := isRunning
	runMut.Unlock()
	if !r {
		return
	}

	startTime := time.Now()
	goalTime := startTime.Unix() + int64(interval.Seconds())
	outputFileName := tempStoragePrefix + startTime.Format(time.RFC3339) + ".avi"

	mut.Lock()
	writer, err := gocv.VideoWriterFile(outputFileName, "MJPG", 55, img.Cols(), img.Rows(), true)
	mut.Unlock()
	if err != nil {
		log.Fatalf("error opening video writer device: %v\n", outputFileName)
	}
	defer writer.Close()

	for {
		curTime := time.Now().Unix()
		if curTime >= goalTime {
			break
		}

		runMut.Lock()
		r := isRunning
		runMut.Unlock()
		if r {
			writer.Write(img)
		}
	}

	log.Printf("%v seconds elapsed; ephemerally written to disk at %v\n", interval.Seconds(), outputFileName)
}

func purgeTemporaryStorage(keepTime time.Duration) {
	for {
		cwd, _ := os.Getwd()
		files, _ := ioutil.ReadDir(cwd)
		for _, f := range files {
			if !f.IsDir() && strings.HasPrefix(f.Name(), tempStoragePrefix) {
				diff := time.Since(f.ModTime())
				if diff >= keepTime {
					os.Remove(f.Name())
					log.Printf("Deleted legacy storage record %v", f.Name())
				}
			}
		}
	}
}
