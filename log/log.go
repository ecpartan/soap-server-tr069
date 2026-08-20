package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/logutils"
	"gopkg.in/natefinch/lumberjack.v2"
)

type CustomLogger struct {
	logger *log.Logger
	filter *logutils.LevelFilter
}

var log_inst = &CustomLogger{}

func InitLogger(wrt io.Writer) {

	log_inst, _ = NewCustomLogger("DEBUG", "app.log")
}
func NewCustomLogger(minLevel logutils.LogLevel, logFile string) (*CustomLogger, error) {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: minLevel,
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   "app.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	filter.Writer = io.MultiWriter(os.Stdout, lumberjackLogger)

	logger := log.New(filter, "", 0)

	return &CustomLogger{
		logger: logger,
		filter: filter,
	}, nil
}

func getArgsString(args ...any) (argsStr string) {
	for _, arg := range args {
		if arg != nil {
			argsStr = argsStr + fmt.Sprintf("'%v', ", arg)
		}
	}
	argsStr = strings.TrimRight(argsStr, ", ")
	return
}

func caller(depth int) string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(depth+1, pc)
	frame, _ := runtime.CallersFrames(pc[:n]).Next()
	idxFile := strings.LastIndexByte(frame.File, '/')
	idx := strings.LastIndexByte(frame.Function, '/')
	idxName := strings.IndexByte(frame.Function[idx+1:], '.') + idx + 1

	return frame.File[idxFile+1:] + ":[" + strconv.Itoa(frame.Line) + "] - " + frame.Function[idxName+1:] + "()"
}

func LogDebug(mes string, args ...any) {
	argsStr := getArgsString(args...) // get formated string with arguments

	if argsStr == "" {
		log_inst.logger.Printf("[DEBUG] %s - %s - %s - %s", time.Now().Format(time.RFC1123), mes, caller(3), mes)
	} else {
		log_inst.logger.Printf("[DEBUG] %s - %s - %s - %s [%s]", time.Now().Format(time.RFC1123), mes, caller(3), mes, argsStr)
	}
}

func LogErr(mes string, args ...any) {
	log_inst.Debugf("[DEBUG]", args...)
}
func (cl *CustomLogger) Debug(msg string, args ...any) {
	cl.logger.Printf("[DEBUG] %s", msg)
}

func (cl *CustomLogger) Info(msg string) {
	cl.logger.Printf("[INFO] %s", msg)
}

func (cl *CustomLogger) Warn(msg string) {
	cl.logger.Printf("[WARN] %s", msg)
}

func (cl *CustomLogger) Error(msg string) {
	cl.logger.Printf("[ERROR] %s", msg)
}

func (cl *CustomLogger) Debugf(format string, args ...interface{}) {
	cl.logger.Printf("[DEBUG] "+format, args...)
}

func (cl *CustomLogger) Infof(format string, args ...interface{}) {
	cl.logger.Printf("[INFO] "+format, args...)
}

func (cl *CustomLogger) Warnf(format string, args ...interface{}) {
	cl.logger.Printf("[WARN] "+format, args...)
}

func (cl *CustomLogger) Errorf(format string, args ...interface{}) {
	cl.logger.Printf("[ERROR] "+format, args...)
}
