package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"

	"gocv.io/x/gocv"
)

func main() {
	defer func() { fmt.Println("shutting down...") }()
	if len(os.Args) < 4 {
		fmt.Println("How to run:\n\tgocam [camera ID] [classifier XML file] [output file]")
		return
	}

	// Parse arguments
	deviceID, _ := strconv.Atoi(os.Args[1])
	xmlFile := os.Args[2]
	saveFile := os.Args[3]

	// Open webcam
	webcam, err := gocv.VideoCaptureDevice(int(deviceID))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer webcam.Close()

	// Open display window
	window := gocv.NewWindow("GoCam")
	defer window.Close()

	// Prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// Color for the rect when faces detected
	blue := color.RGBA{0, 0, 255, 0}

	// Initializes the image
	if ok := webcam.Read(&img); !ok {
		fmt.Printf("Cannot read device %v\n", deviceID)
		return
	}

	// Open stream to write a file to
	writer, err := gocv.VideoWriterFile(saveFile, "MJPG", 10, img.Cols(), img.Rows(), true)
	if err != nil {
		fmt.Printf("error opening video writer device: %v\n", saveFile)
		fmt.Println(err)
		return
	}
	defer writer.Close()

	// Load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load(xmlFile) {
		fmt.Printf("Error reading cascade file: %v\n", xmlFile)
		return
	}

	fmt.Printf("start reading camera device: %v\n", deviceID)
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("cannot read device %d\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

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

		// Write the data to file
		writer.Write(img)

		// Show the image in the window and wait 1 millisecond
		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}
