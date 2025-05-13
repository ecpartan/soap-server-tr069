package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/hashicorp/logutils"
)

// logFilter represent a custom logger seting
var logFilter = &logutils.LevelFilter{
	Levels:   []logutils.LogLevel{"DEBUG", "INFO", "ERROR", "CRITICAL"},
	MinLevel: logutils.LogLevel("INFO"),
	Writer:   os.Stdout,
}

// InitLogger init custom logger
func InitLogger(wrt io.Writer) {
	logFilter.Writer = wrt
	log.SetOutput(logFilter)
}

// SetFilter set log level
func SetLevel(lev string) {
	logFilter.SetMinLevel(logutils.LogLevel(lev))
}

// caller return caller function name
func LogDebug(mes string, args ...interface{}) {
	printMsg("[DEBUG]", 0, mes, args...)
}

func printMsg(level string, depth int, mes string, args ...interface{}) {
	// Chek for appropriate level of logging
	if logFilter.Check([]byte(level)) {
		argsStr := getArgsString(args...) // get formated string with arguments

		if argsStr == "" {
			log.Printf("%s - %s - %s", level, caller(depth+3), mes)
		} else {
			log.Printf("%s - %s - %s [%s]", level, caller(depth+3), mes, argsStr)
		}
	}
}

// getArgsString return formated string with arguments
func getArgsString(args ...interface{}) (argsStr string) {
	for _, arg := range args {
		if arg != nil {
			argsStr = argsStr + fmt.Sprintf("'%v', ", arg)
		}
	}
	argsStr = strings.TrimRight(argsStr, ", ")
	return
}

// caller returns a Valuer that returns a file and line from a specified depth in the callstack.
func caller(depth int) string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(depth+1, pc)
	frame, _ := runtime.CallersFrames(pc[:n]).Next()
	idxFile := strings.LastIndexByte(frame.File, '/')
	idx := strings.LastIndexByte(frame.Function, '/')
	idxName := strings.IndexByte(frame.Function[idx+1:], '.') + idx + 1

	return frame.File[idxFile+1:] + ":[" + strconv.Itoa(frame.Line) + "] - " + frame.Function[idxName+1:] + "()"
}
