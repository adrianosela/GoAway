package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"gocv.io/x/gocv"
)

const (
	minDiffArea = 3000

	// DetectorStatusReady is the status of the detector when it
	// has completed initialization
	DetectorStatusReady = "Ready"

	// DetectorStatusMotionDetected is the status of the detector when
	// it detects a difference between two consecurive frames on the recording
	// device (i.e. motion is detected)
	DetectorStatusMotionDetected = "Motion Detected"

	// DetectorStatusClosed is the status of the detector when it is no
	// longer being closed, once it is closed, it cannot be re-initialized
	DetectorStatusClosed = "Closed"
)

var (
	boundingRectColor         = color.RGBA{0, 0, 255, 0}
	statusReadyColor          = color.RGBA{0, 255, 0, 0}
	statusMotionDetectedColor = color.RGBA{255, 0, 0, 0}
)

// Detector is an abstraction for a motion detector
type Detector struct {
	camera        *gocv.VideoCapture
	window        *gocv.Window
	baseImgMatrix gocv.Mat
	diffMatrix    gocv.Mat
	threshMatrix  gocv.Mat
	bgSubtractor  gocv.BackgroundSubtractorMOG2
	statusColor   color.RGBA
	status        string
	onDetect      func()
}

func (d *Detector) waitForNextFrame() error {
	for {
		if ok := d.camera.Read(&d.baseImgMatrix); !ok {
			return fmt.Errorf("Video Device Closed")
		}
		if !d.baseImgMatrix.Empty() {
			d.status, d.statusColor  = DetectorStatusReady, statusReadyColor
			return nil
		}
	}
}

func (d *Detector) prepareCurFrame() {
	// foreground (diff matrix) = curFrame - prevFrame
	d.bgSubtractor.Apply(d.baseImgMatrix, &d.diffMatrix)
	// distinguish items in foreground by distinguishing differences > 25 in RGB vals of img
	gocv.Threshold(d.diffMatrix, &d.threshMatrix, 25, 255, gocv.ThresholdBinary)
	// transformation that produces an image that is the same shape as the
	// original, but is a different size
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernel.Close()
	gocv.Dilate(d.threshMatrix, &d.threshMatrix, kernel)
}

func (d *Detector) findAndDrawContours() {
	contours := gocv.FindContours(d.threshMatrix, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	for i, c := range contours {
		area := gocv.ContourArea(c)
		if area < minDiffArea {
			continue
		}
		d.status, d.statusColor = DetectorStatusMotionDetected, statusMotionDetectedColor
		// run user provided on-detect function
		if d.onDetect != nil {
			go d.onDetect()
		}
		gocv.DrawContours(&d.baseImgMatrix, contours, i, d.statusColor, 2)
		rect := gocv.BoundingRect(c)
		gocv.Rectangle(&d.baseImgMatrix, rect, boundingRectColor, 2)
	}
}

func (d *Detector) displayResult() bool {
	gocv.PutText(&d.baseImgMatrix, d.status, image.Pt(10, 20), gocv.FontHersheyPlain, 1.2, d.statusColor, 2)
	d.window.IMShow(d.baseImgMatrix)
	// are we finished?
	return d.window.WaitKey(1) == 27
}

// NewMotionDetector is the constructor for a Detector
func NewMotionDetector(camID int, winTitle string, onDetect func()) (*Detector, error) {
	cam, err := gocv.OpenVideoCapture(camID)
	if err != nil {
		return nil, err
	}
	return &Detector{
		camera:        cam,
		window:        gocv.NewWindow(winTitle),
		baseImgMatrix: gocv.NewMat(),
		diffMatrix:    gocv.NewMat(),
		threshMatrix:  gocv.NewMat(),
		bgSubtractor:  gocv.NewBackgroundSubtractorMOG2(),
		statusColor:   statusReadyColor,
		status:        DetectorStatusReady,
		onDetect: onDetect,
	}, nil
}

// Start initializes the motion detector
func (d *Detector) Start() error {
	for {
		if err := d.waitForNextFrame(); err != nil {
			return err
		}
		d.prepareCurFrame()
		d.findAndDrawContours()
		if done := d.displayResult(); done {
			break
		}
	}
	return nil
}

// Close handles closing gocv resources
func (d *Detector) Close() {
	d.status = DetectorStatusClosed
	if err := d.camera.Close(); err != nil {
		log.Printf("could not close camera: %s", err)
	}
	if err := d.window.Close(); err != nil {
		log.Printf("could not close window: %s", err)
	}
	if err := d.baseImgMatrix.Close(); err != nil {
		log.Printf("could not close image matrix: %s", err)
	}
	if err := d.diffMatrix.Close(); err != nil {
		log.Printf("could not close diff matrix: %s", err)
	}
	if err := d.threshMatrix.Close(); err != nil {
		log.Printf("could not close threshold matrix: %s", err)
	}
	if err := d.bgSubtractor.Close(); err != nil {
		log.Printf("could not close background subtractor: %s", err)
	}
}

func main() {
	detector, err := NewMotionDetector(0, "Motion Detector", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer detector.Close()
	log.Fatal(detector.Start())
}