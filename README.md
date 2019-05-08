# GoAway - An OpenCV based motion detector written in Go

[![Go Report Card](https://goreportcard.com/badge/github.com/adrianosela/GoAway)](https://goreportcard.com/report/github.com/adrianosela/GoAway)
[![Documentation](https://godoc.org/github.com/adrianosela/GoAway?status.svg)](https://godoc.org/github.com/adrianosela/GoAway)
[![GitHub issues](https://img.shields.io/github/issues/adrianosela/GoAway.svg)](https://github.com/adrianosela/GoAway/issues)
[![license](https://img.shields.io/github/license/adrianosela/goaway.svg)](https://github.com/adrianosela/GoAway/blob/master/LICENSE)

**Complete examples can be found in the /examples subdirectory of this repository**

### Simple Usage

```
package main

import (
	"github.com/adrianosela/GoAway/detector"
)

func main() {
	md, err := detector.NewMotionDetector(0, "Motion Detector", nil)
	defer md.Close()
	if err != nil { /* handle error */ }
	md.Start()
}
```

### With On-Detect Function

```
import (
	"github.com/adrianosela/GoAway/detector"
)

func main() {
	md, err := detector.NewMotionDetector(0, "Motion Detector", func() {
		// do this whenever motion is detected
		// e.g. log, send yourself an email, etc...
	})
	defer md.Close()
	if err != nil { /* handle error */ }
	md.Start()
}
```