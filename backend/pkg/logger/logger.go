package logger

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		infoLog:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		errorLog: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
		debugLog: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
	}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	l.infoLog.Println(fmt.Sprintf(msg, args...))
}

func (l *Logger) Error(msg string, args ...interface{}) {
	l.errorLog.Println(fmt.Sprintf(msg, args...))
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	l.debugLog.Println(fmt.Sprintf(msg, args...))
}
