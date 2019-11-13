package logging

import (
	"errors"
	"fmt"
	"log"
	"os"
)

var validLogLevels = []string{"ERROR", "WARNING", "INFO", "DEBUG"}

type Loggers struct {
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

func LoggerSetUp() (*Loggers, error) {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}
	l := Loggers{}
	l.Error = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	if isValidLogLevel(logLevel) {
		switch logLevel {
		case "DEBUG":
			l.Debug = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
			l.Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
			l.Warning = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
		case "INFO":
			l.Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
			l.Warning = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
		case "WARNING":
			l.Warning = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
		}
		return &l, nil
	} else {
		return nil, errors.New(fmt.Sprintf("An invalid log level was provided. Accepted values are %v", validLogLevels))
	}
}

func isValidLogLevel(l string) bool {
	return stringInSlice(l, validLogLevels)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
