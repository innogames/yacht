package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

var logger struct {
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

func init_logger(app_state *AppState) {
	var DebugWriter io.Writer

	if app_state.verbose {
		DebugWriter = os.Stderr
	} else {
		DebugWriter = ioutil.Discard
	}
	logger.Debug = log.New(DebugWriter, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	logger.Warning = log.New(os.Stderr, "WARNING: ", log.Ldate|log.Ltime)
	logger.Error = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
