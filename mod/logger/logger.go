package logger

import (
	"fmt"
	"os"
)

var cReset = "\033[0m"
var cError = "\033[31m"
var cInfo = "\033[32m"
var cLog = "\033[36m"
var cSub = "\033[33m"

func Info(source string, exec string, messages ...interface{}) {
	fmt.Fprintf(os.Stdout, cInfo+"%s ", source)
	fmt.Fprintf(os.Stdout, cSub+"%s ", exec)
	fmt.Fprintf(os.Stdout, cLog+"%v\n"+cReset, messages)
}

func Error(source string, exec string, messages ...interface{}) {
	fmt.Fprintf(os.Stderr, cError+"%s "+cReset, source)
	fmt.Fprintf(os.Stderr, cError+"%s "+cReset, exec)
	fmt.Fprintf(os.Stderr, cError+"%v\n"+cReset, messages)
}
