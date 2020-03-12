package main

// 31.08.2015 created by Formeo

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/pharmacy72/gobak/application"
	"github.com/pharmacy72/gobak/config"
	"github.com/pharmacy72/gobak/logger"
	"github.com/pharmacy72/gobak/smail"
	"os"
)

var app *application.Application

func main() {
	defer func() {
		if e := recover(); e != nil {
			if app == nil || !app.Verbose {
				if !app.ErrOut2mail(e) {
					fmt.Fprintln(os.Stderr, e)
				}
			} else if app.Verbose {
				fmt.Fprintln(os.Stderr, e.(error))
			}
			os.Exit(1)
		}
	}()
	conf := config.Current()
	err := logger.InitSentry(conf)
	if err != nil {
		panic(err)
	}
	log, err := logger.NewLogger(conf)
	if err != nil {
		sentry.CaptureException(err)
	}
	sMail:=smail.NewMailApp("smtpServerUrl", log, "emailTo string", "emailFrom string")
	app = application.NewApplication(sMail)
	app.Run()
	return

}
