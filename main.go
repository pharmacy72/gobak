package main

// 31.08.2015 created by Formeo

//TODO: Logger
import (
	"fmt"
	"github.com/pharmacy72/gobak/application"
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
	app = application.NewApplication()
	app.Run()
	return

}
