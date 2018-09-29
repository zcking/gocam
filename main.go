package main

import (
	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const TempStoragePrefix string = "TMP_"

var (
	deviceID int
	err      error
	webcam   *gocv.VideoCapture
	stream   *mjpeg.Stream
	xmlFile  string
	classifier gocv.CascadeClassifier
	blue 	 color.RGBA
)

var (
	img 	gocv.Mat
	mut 	sync.Mutex
)

var (
	isRunning = false
	runMut 	  sync.Mutex
)

var writeMut sync.Mutex

func main() {
	defer func() { log.Println("Gocam shutting down...") }()
	if len(os.Args) < 6 {
		log.Println("How to run:\n\tgocam [camera ID] [classifier XML file] [host] [temp recording length] [temp keep time]")
		return
	}

	// Parse arguments
	deviceID, err = strconv.Atoi(os.Args[1])
	xmlFile = os.Args[2]
	host := os.Args[3]
	tempRecLength, _ := time.ParseDuration(os.Args[4])
	tempKeepTime, _ := time.ParseDuration(os.Args[5])

	// Color for the rect when faces detected
	blue = color.RGBA{B: 255}

	// Open webcam
	webcam, err = gocv.VideoCaptureDevice(int(deviceID))
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer webcam.Close()

	// Prepare image matrix
	img = gocv.NewMat()
	defer img.Close()

	// Create the mjpeg stream
	stream = mjpeg.NewStream()

	// Enable face detection
	// Load classifier to recognize faces
	classifier = gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load(xmlFile) {
		log.Fatalf("Error reading cascade file: %v\n", xmlFile)
		return
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
				detectFaces()
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
	go func() {
		for {
			writeMut.Lock()
			writeTemporaryStorage(tempRecLength)
			writeMut.Unlock()
		}
	}()
	// Purge any older temporary files (beyond the keep time)
	go purgeTemporaryStorage(tempKeepTime)

	// Spin up the controller server
	http.HandleFunc("/health", func(w http.ResponseWriter, request *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/api/power/off", func(w http.ResponseWriter, request *http.Request) {
		runMut.Lock()
		defer runMut.Unlock()

		// Shutdown the camera
		isRunning = false
		log.Println("Gocam powering off...")
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/api/power/on", func(w http.ResponseWriter, request *http.Request) {
		runMut.Lock()
		defer runMut.Unlock()

		// Poweron the camera
		isRunning = true
		log.Printf("Gocam powering on... Streaming to %v\n", host)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/cam", func(w http.ResponseWriter, request *http.Request) {
		runMut.Lock()
		r := isRunning
		runMut.Unlock()
		if !r {
			w.WriteHeader(http.StatusNotFound)
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
	outputFileName := TempStoragePrefix + startTime.Format(time.RFC3339) + ".avi"

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
			if !f.IsDir() && strings.HasPrefix(f.Name(), TempStoragePrefix) {
				diff := time.Since(f.ModTime())
				if diff >= keepTime {
					os.Remove(f.Name())
					log.Printf("Deleted legacy storage record %v", f.Name())
				}
			}
		}
	}
}