package logger

import "io"

// Logger defines the behavior of a data logger component
type Logger interface {
	Trace(message string, args ...interface{})
	Debug(message string, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message string, args ...interface{})
	LogIfError(err error, args ...interface{})
	SetLevel(logLevel LogLevel)
	IsInterfaceNil() bool
}

// LogLineHandler defines the get methods for a log line struct used by the formatter interface
type LogLineHandler interface {
	GetMessage() string
	GetLogLevel() int32
	GetArgs() []string
	GetTimestamp() int64
	IsInterfaceNil() bool
}

// Formatter describes what a log formatter should be able to do
type Formatter interface {
	Output(line LogLineHandler) []byte
	IsInterfaceNil() bool
}

// LogOutputHandler defines the properties of a subject-observer component
// able to output log lines
type LogOutputHandler interface {
	Output(line *LogLine)
	AddObserver(w io.Writer, format Formatter) error
	RemoveObserver(w io.Writer) error
	ClearObservers()
	IsInterfaceNil() bool
}
