/*
 *
 * Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */

package nebula_go

import (
	"log"
)

var (
	_ Logger = (*DefaultLogger)(nil)
	_ Logger = (*NoLogger)(nil)
)

// Logger interface.
type Logger interface {
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
}

// DefaultLogger type.
type DefaultLogger struct{}

// Info method.
func (l DefaultLogger) Info(msg string) {
	log.Printf("[INFO] %s\n", msg)
}

// Warn method.
func (l DefaultLogger) Warn(msg string) {
	log.Printf("[WARNING] %s\n", msg)
}

// Error method.
func (l DefaultLogger) Error(msg string) {
	log.Printf("[ERROR] %s\n", msg)
}

// Fatal method.
func (l DefaultLogger) Fatal(msg string) {
	log.Fatalf("[FATAL] %s\n", msg)
}

// NoLogger type. Will omit all messages except Fatal.
type NoLogger struct{}

// Info method.
func (l NoLogger) Info(msg string) {}

// Warn method.
func (l NoLogger) Warn(msg string) {}

// Error method.
func (l NoLogger) Error(msg string) {}

// Fatal method.
func (l NoLogger) Fatal(msg string) {
	log.Fatalf("[FATAL] %s\n", msg)
}
