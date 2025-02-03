// logger/logger.go
package logger

import (
    "github.com/sirupsen/logrus"
)

// Logger is a globally accessible logger instance.
var Logger = logrus.New()

func init() {
    SetFormatter("text") // Default formatter
    SetLevel("info")     // Default level
}

// SetFormatter allows dynamic configuration of the log formatter.
func SetFormatter(format string) {
    switch format {
    case "json":
        Logger.SetFormatter(&logrus.JSONFormatter{})
    default:
        Logger.SetFormatter(&logrus.TextFormatter{
            FullTimestamp: true,
        })
    }
}

// SetLevel sets the logging level based on a string input.
func SetLevel(level string) {
    switch level {
    case "debug":
        Logger.SetLevel(logrus.DebugLevel)
    case "warn":
        Logger.SetLevel(logrus.WarnLevel)
    case "error":
        Logger.SetLevel(logrus.ErrorLevel)
    default:
        Logger.SetLevel(logrus.InfoLevel)
    }
}
