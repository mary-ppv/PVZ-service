package logger

import (
	"log"
	"os"
)

var Log *log.Logger

func Init() {
	Log = log.New(os.Stdout, "[PVZ] ", log.LstdFlags|log.Lshortfile)
}
