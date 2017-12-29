package lemon

import (
	"log"
	"os"
	"fmt"
)

type Logger interface {
	Debug(s ...interface{})
}

func newLogger() *logger {
	return &logger{log: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)}
}

type logger struct {
	log *log.Logger
}

func (l *logger) Debug(s ...interface{}) {
	l.log.Output(3, fmt.Sprintln(s...))
}
