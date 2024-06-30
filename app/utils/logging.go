package utils

import (
	"io"
	"log"
	"os"

	gommon_log "github.com/labstack/gommon/log"
)

func LoggingSettings(logFile string) {
	logfile, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	multiLogFile := io.MultiWriter(os.Stdout, logfile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(multiLogFile)
}

func GetFilePointer(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
}

func GetLEVEL(v string) gommon_log.Lvl {

	switch v {
	case "DEBUG":
		return gommon_log.DEBUG
	case "INFO":
		return gommon_log.INFO
	case "WARN":
		return gommon_log.WARN
	case "ERROR":
		return gommon_log.ERROR
	case "OFF":
		return gommon_log.OFF
	default:
		return gommon_log.DEBUG
	}
}
