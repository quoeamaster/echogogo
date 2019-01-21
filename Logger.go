package main

import (
	"bytes"
	"fmt"
	"time"
)

const LogLevelTrace = 0
const LogLevelDebug = 1
const LogLevelInfo = 2
const LogLevelWarning = 3
const LogLevelError = 4

type LogConfig struct {
	DefaultLevel    int
	Filename        string	// e.g. Server.go or Main.go etc (should remove the suffix ".go" though)
	DefaultFuncName string // if func name is not provided, will use this value to build the log line
}

type Logger struct {
	ThresholdLogLevel int	// the threshold for logging (e.g. info; which means all logs lower than INFO would be skipped)
}

func NewLogger(thresholdLogLevel int) Logger {
	ptr := new(Logger)
	if thresholdLogLevel >= 0 && thresholdLogLevel <= 4 {
		ptr.ThresholdLogLevel = thresholdLogLevel
	} else {
		ptr.ThresholdLogLevel = LogLevelInfo
	}
	return *ptr
}

// logging method with all parameters required
func (l *Logger) Log(message string, logLevel int, filename string, funcName string) (charsLogged int, err error)  {
	if logLevel < l.ThresholdLogLevel {
		return 0, nil
	}
	var buffer bytes.Buffer

	buffer.WriteString("[")
	buffer.WriteString(_getTimeNow().String())
	buffer.WriteString("][")
	// log level
	buffer.WriteString(_translateLogLevelString(logLevel))
	buffer.WriteString("][")
	// filename.Funcname
	if len(filename) > 0 {
		buffer.WriteString(filename)
	} else {
		buffer.WriteString("-")
	}
	buffer.WriteString(".")
	if len(funcName) >0 {
		buffer.WriteString(funcName)
	} else {
		buffer.WriteString("-")
	}
	buffer.WriteString("] ")
	buffer.WriteString(message)

	charsLogged, err = fmt.Printf("%v\n", buffer.String())

	return
}

// simple log method
func (l *Logger) LogWithFuncName(message string, funcName string, logConfig ...LogConfig) (charsLogged int, err error)  {
	isConfigValid := logConfig != nil && len(logConfig) >= 0
	var buffer bytes.Buffer

	buffer.WriteString("[")
	buffer.WriteString(time.Now().UTC().String())
	buffer.WriteString("][")
	// log level
	if isConfigValid {
		buffer.WriteString(_translateLogLevelString(logConfig[0].DefaultLevel))
	} else {
		buffer.WriteString("trace")
	}
	buffer.WriteString("][")
	// filename.Funcname
	if isConfigValid {
		buffer.WriteString(logConfig[0].Filename)
	} else {
		buffer.WriteString(" - ")
	}
	buffer.WriteString(".")

	if len(funcName) > 0 {
		buffer.WriteString(funcName)
	} else {
		if isConfigValid && len(logConfig[0].DefaultFuncName) > 0 {
			buffer.WriteString(logConfig[0].DefaultFuncName)
		} else {
			buffer.WriteString(" - ")
		}
	}
	buffer.WriteString("] ")
	buffer.WriteString(message)

	charsLogged, err = fmt.Printf("%v\n", buffer.String())

	return
}


// method to get the current time (implementation varies)
func _getTimeNow() time.Time {
	return time.Now().UTC()
}

// lookup method to translate the logLevel (int) back to a string
func _translateLogLevelString(logLevel int) string {
	switch logLevel {
	case 0:
		return "trace"
	case 1:
		return "debug"
	case 2:
		return "info"
	case 3:
		return "WARNING"
	case 4:
		return "ERROR"
	default:
		return "trace"
	}
}