package main

import (
	"fmt"
	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"syscall"
)

var (
	deviceID int
	err      error
	webcam   *gocv.VideoCapture
	stream   *mjpeg.Stream
	saveFile string
	xmlFile  string
	classifier gocv.CascadeClassifier
	blue 	 color.RGBA
)

var (
	img 	gocv.Mat
	mut 	sync.Mutex
)

func main() {
	defer func() { fmt.Println("shutting down...") }()
	if len(os.Args) < 4 {
		fmt.Println("How to run:\n\tgocam [camera ID] [classifier XML file] [host] [output file]")
		return
	}

	// Parse arguments
	deviceID, err = strconv.Atoi(os.Args[1])
	xmlFile = os.Args[2]
	host := os.Args[3]
	saveFile = os.Args[4]

	// Color for the rect when faces detected
	blue = color.RGBA{0, 0, 255, 0}

	// Open webcam
	webcam, err = gocv.VideoCaptureDevice(int(deviceID))
	if err != nil {
		fmt.Println(err)
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
		fmt.Printf("Error reading cascade file: %v\n", xmlFile)
		return
	}

	// Capture a single image just to initialize the image variable
	captureImage()

	// Capture images from the camera in parallel
	go func() {
		for {
			captureImage()
			detectFaces()
		}
	}()

	// Start capturing for mjpeg stream
	go mjpegCapture()
	fmt.Println("Streaming to " + host)

	// Start HTTP Server
	http.Handle("/", stream)
	log.Fatal(http.ListenAndServe(host, nil))
}

func detectFaces() {
	mut.Lock()
	defer mut.Unlock()

	// Detect faces
	rects := classifier.DetectMultiScale(img)
	//fmt.Printf("found %d faces\n", len(rects))

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
		fmt.Printf("Device closed: %v\n", deviceID)
		syscall.Exit(-1)
	}
	if img.Empty() {
		syscall.Exit(-1)
	}
}

func mjpegCapture() {
	for {
		mut.Lock()
		buf, _ := gocv.IMEncode(".jpg", img)
		stream.UpdateJPEG(buf)
		mut.Unlock()
	}
}