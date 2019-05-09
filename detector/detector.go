package detector

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"gocv.io/x/gocv"
)

const (
	escapeKey = 27

	// NotSensitive represents a large mimumum diff contour area of an image
	// for a minimum sensitivity (not very sensitive) motion detector
	NotSensitive = 9000

	// DefaultSensitive represents a medium minimum diff contour area of an image
	// for a medium sensitivity (default sensitive) motion detector
	DefaultSensitive = 6000

	// VerySensitive represents a small minimum diff contour area of an image
	// for a maximum sensitivity (very sensitive) motion detector
	VerySensitive = 3000

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
	boundingRectColor         = color.RGBA{0, 0, 0, 0}       // black
	statusReadyColor          = color.RGBA{255, 255, 255, 0} // white
	statusMotionDetectedColor = color.RGBA{255, 0, 255, 0}   // purple
)

// Detector is an abstraction for a motion detector
type Detector struct {
	camera             *gocv.VideoCapture
	window             *gocv.Window
	baseImgMatrix      gocv.Mat
	diffMatrix         gocv.Mat
	threshMatrix       gocv.Mat
	bgSubtractor       gocv.BackgroundSubtractorMOG2
	statusColor        color.RGBA
	minDiffContourArea float64
	status             string
	onDetect           func()
}

func (d *Detector) waitForNextFrame() error {
	for {
		if ok := d.camera.Read(&d.baseImgMatrix); !ok {
			return fmt.Errorf("Video Device Closed")
		}
		if !d.baseImgMatrix.Empty() {
			d.status, d.statusColor = DetectorStatusReady, statusReadyColor
			return nil
		}
	}
}

func (d *Detector) prepareCurrentFrame() {
	// foreground (diff matrix) = curFrame - prevFrame
	d.bgSubtractor.Apply(d.baseImgMatrix, &d.diffMatrix)
	// get rid of pixels with too small or too large values
	gocv.Threshold(d.diffMatrix, &d.threshMatrix, 25, 255, gocv.ThresholdBinary)
	// Dilate: transformation that produces an image that is the same shape as the
	// original, but is a different size
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernel.Close()
	gocv.Dilate(d.threshMatrix, &d.threshMatrix, kernel)
}

func (d *Detector) findAndDrawContours() {
	contours := gocv.FindContours(d.threshMatrix, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	for i, c := range contours {
		area := gocv.ContourArea(c)
		if area < d.minDiffContourArea {
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
	return d.window.WaitKey(1) == escapeKey
}

// NewMotionDetector is the constructor for a Detector
func NewMotionDetector(camID int, winTitle string, onDetect func()) (*Detector, error) {
	cam, err := gocv.OpenVideoCapture(camID)
	if err != nil {
		return nil, err
	}
	return &Detector{
		camera:             cam,
		window:             gocv.NewWindow(winTitle),
		baseImgMatrix:      gocv.NewMat(),
		diffMatrix:         gocv.NewMat(),
		threshMatrix:       gocv.NewMat(),
		bgSubtractor:       gocv.NewBackgroundSubtractorMOG2(),
		statusColor:        statusReadyColor,
		status:             DetectorStatusReady,
		onDetect:           onDetect,
		minDiffContourArea: NotSensitive,
	}, nil
}

// Start initializes the motion detector
func (d *Detector) Start() error {
	for {
		if err := d.waitForNextFrame(); err != nil {
			return err
		}
		d.prepareCurrentFrame()
		d.findAndDrawContours()
		if done := d.displayResult(); done {
			break
		}
	}
	return nil
}

// Status returns the status of the detector
func (d *Detector) Status() string {
	return d.status
}

// SnapshotJPG returns a jpg encoded byte slice containing
// the latest image taken from the video capture device
func (d *Detector) SnapshotJPG() ([]byte, error) {
	return gocv.IMEncode(".jpg", d.baseImgMatrix)
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
