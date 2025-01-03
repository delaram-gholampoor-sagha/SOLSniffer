package logger

import (
	"fmt"
	"os"

	"github.com/delaram-gholampoor-sagha/SOLSniffer/configs"
	"github.com/sirupsen/logrus"
)

// Logger is the global logger instance
var Logger *logrus.Logger

// Register initializes the logging package based on the provided AppConfig
func Register(appConfig configs.AppConfig) error {
	Logger = logrus.New()

	Logger.Out = os.Stdout

	level, err := logrus.ParseLevel(appConfig.Log.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %s", appConfig.Log.LogLevel)
	}
	Logger.SetLevel(level)

	if appConfig.Log.PrettyPrint {
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	} else {
		Logger.SetFormatter(&logrus.JSONFormatter{})
	}

	return nil
}

// Info logs an info message
func Info(message string, fields map[string]interface{}) {
	Logger.WithFields(fields).Info(message)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

// Debug logs a debug message
func Debug(message string, fields map[string]interface{}) {
	Logger.WithFields(fields).Debug(message)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

// Warn logs a warning message
func Warn(message string, fields map[string]interface{}) {
	Logger.WithFields(fields).Warn(message)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

// Error logs an error message
func Error(message string, fields map[string]interface{}) {
	Logger.WithFields(fields).Error(message)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

// Fatal logs a fatal error message and exits the application
func Fatal(message string, fields map[string]interface{}) {
	Logger.WithFields(fields).Fatal(message)
}

// Fatalf logs a formatted fatal error message and exits the application
func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

// Panic logs a panic message and panics
func Panic(message string, fields map[string]interface{}) {
	Logger.WithFields(fields).Panic(message)
}

// Panicf logs a formatted panic message and panics
func Panicf(format string, args ...interface{}) {
	Logger.Panicf(format, args...)
}
