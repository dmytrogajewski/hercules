package main

import (
	"os"
	"testing"

	"github.com/dmytrogajewski/hercules/internal/app/core"
	"github.com/stretchr/testify/assert"
)

// TestLoggerBehavior tests basic logger behavior
func TestLoggerBehavior(t *testing.T) {
	// Test NoOpLogger doesn't output anything
	noOpLogger := &core.NoOpLogger{}
	noOpLogger.Info("test")
	noOpLogger.Warn("test")
	noOpLogger.Error("test")
	noOpLogger.Infof("test %d", 123)
	noOpLogger.Warnf("test %s", "warning")
	noOpLogger.Errorf("test %v", "error")
	// No assertions needed - NoOpLogger should not panic or output anything

	// Test SlogLogger can be created
	slogLogger := core.NewSlogLogger(os.Stdout)
	assert.NotNil(t, slogLogger, "SlogLogger should be created successfully")

	// Test FileLogger can be created with a file
	tmpfile, err := os.CreateTemp("", "hercules_test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	fileLogger := core.NewFileLogger(tmpfile)
	assert.NotNil(t, fileLogger, "FileLogger should be created successfully")
}
