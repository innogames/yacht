package logger

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

var Debug *log.Logger
var Info *log.Logger
var Warning *log.Logger
var Error *log.Logger

func InitLoggers(verbose bool) {
	var DebugWriter io.Writer

	if verbose {
		DebugWriter = os.Stderr
	} else {
		DebugWriter = ioutil.Discard
	}

	Debug = log.New(DebugWriter, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	Warning = log.New(os.Stderr, "WARNING: ", log.Ldate|log.Ltime)
	Error = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
