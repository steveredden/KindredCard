/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	TRACE LogLevel = iota // Most verbose - extremely detailed
	DEBUG                 // Detailed debugging
	INFO                  // General information
	WARN                  // Warnings
	ERROR                 // Errors
	FATAL                 // Critical errors that stop execution
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger is a custom logger with level support
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

var defaultLogger *Logger

// Init initializes the logger from environment variable
// Reads LOG_LEVEL env var (DEBUG, INFO, WARN, ERROR, FATAL)
// Defaults to INFO if not set or invalid
func Init() {
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "INFO" // Default level
	}

	level := parseLogLevel(levelStr)
	defaultLogger = &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}

	defaultLogger.Info("Logger initialized with level: %s", level.String())
}

// parseLogLevel converts a string to LogLevel
func parseLogLevel(s string) LogLevel {
	switch strings.ToUpper(s) {
	case "TRACE":
		return TRACE
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO // Default to INFO for invalid values
	}
}

// SetLevel changes the current log level
func SetLevel(level LogLevel) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.level = level
}

// GetLevel returns the current log level
func GetLevel() LogLevel {
	if defaultLogger == nil {
		Init()
	}
	return defaultLogger.level
}

// shouldLog checks if a message at the given level should be logged
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

// log formats and outputs a log message with the given level
func (l *Logger) log(level LogLevel, format string, v ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	prefix := fmt.Sprintf("[%s] ", level.String())
	message := fmt.Sprintf(format, v...)
	l.logger.Printf("%s%s", prefix, message)
}

// Trace logs a trace message (most verbose)
func (l *Logger) Trace(format string, v ...interface{}) {
	l.log(TRACE, format, v...)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(DEBUG, format, v...)
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.log(INFO, format, v...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	l.log(WARN, format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.log(ERROR, format, v...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log(FATAL, format, v...)
	os.Exit(1)
}

// Package-level convenience functions

// Trace logs a trace message using the default logger
func Trace(format string, v ...interface{}) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.Trace(format, v...)
}

// Debug logs a debug message using the default logger
func Debug(format string, v ...interface{}) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.Debug(format, v...)
}

// Info logs an info message using the default logger
func Info(format string, v ...interface{}) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.Info(format, v...)
}

// Warn logs a warning message using the default logger
func Warn(format string, v ...interface{}) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.Warn(format, v...)
}

// Error logs an error message using the default logger
func Error(format string, v ...interface{}) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.Error(format, v...)
}

// Fatal logs a fatal message using the default logger and exits
func Fatal(format string, v ...interface{}) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.Fatal(format, v...)
}

// Printf provides compatibility with standard log.Printf
func Printf(format string, v ...interface{}) {
	Info(format, v...)
}
