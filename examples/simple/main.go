package main

import (
	"github.com/adrianosela/GoAway/detector"
	"log"
)

func main() {
	md, err := detector.NewMotionDetector(0, "Motion Detector", func(){
		// do this whenever motion is detected
		// e.g. log, send yourself an email, etc...
	})
	defer md.Close()
	if err != nil {
		log.Fatal(err)
	}
	md.Start()
}
