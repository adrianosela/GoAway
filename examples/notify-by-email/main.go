package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"time"

	"github.com/adrianosela/GoAway/detector"
)

var (
	// initialize time to "fifteen seconds ago"
	lastMail = time.Now().Add(-15 * time.Second)
)

func notifyMeByEmail(snapshot []byte) error {
	from := os.Getenv("GMAIL_USER")
	pass := os.Getenv("GMAIL_PASS")
	recipients := []string{from}
	smtpHost := "smtp.gmail.com"
	smtpPort := 587
	smtpAddr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)
	smtpAuth := smtp.PlainAuth("", from, pass, smtpHost)
	return smtp.SendMail(smtpAddr, smtpAuth, from, recipients, snapshot)
}

func main() {

	onDetect := func() {
		// only send at most one email per 15 seconds
		if time.Now().After(lastMail.Add(15 * time.Second)) {
			lastMail = time.Now()
			notifyMeByEmail([]byte("Motion has been detected in the room"))
		}
	}

	md, err := detector.NewMotionDetector(0, "Motion Detector", onDetect)
	if err != nil {
		log.Fatal(err)
	}

	defer md.Close()

	md.Start()
}
