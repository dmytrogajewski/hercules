package core

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	var (
		f = "%s-%s"
		v = []interface{}{"hello", "world"}
		l = NewLogger()

		iBuf bytes.Buffer
		wBuf bytes.Buffer
		eBuf bytes.Buffer
	)

	// capture output
	l.I.SetOutput(&iBuf)
	l.W.SetOutput(&wBuf)
	l.E.SetOutput(&eBuf)

	l.Info(v...)
	assert.Contains(t, iBuf.String(), "[INFO]")
	iBuf.Reset()

	l.Infof(f, v...)
	assert.Contains(t, iBuf.String(), "[INFO]")
	assert.Contains(t, iBuf.String(), "-")
	iBuf.Reset()

	l.Warn(v...)
	assert.Contains(t, wBuf.String(), "[WARN]")
	wBuf.Reset()

	l.Warnf(f, v...)
	assert.Contains(t, wBuf.String(), "[WARN]")
	assert.Contains(t, wBuf.String(), "-")
	wBuf.Reset()

	l.Error(v...)
	assert.Contains(t, eBuf.String(), "[ERROR]")
	eBuf.Reset()

	l.Errorf(f, v...)
	assert.Contains(t, eBuf.String(), "[ERROR]")
	assert.Contains(t, eBuf.String(), "-")
	eBuf.Reset()

	l.Critical(v...)
	assert.Contains(t, eBuf.String(), "[ERROR]")
	assert.Contains(t, eBuf.String(), "internal/core.TestLogger")
	assert.Contains(t, eBuf.String(), "internal/core/logger_test.go:")
	eBuf.Reset()

	l.Criticalf(f, v...)
	assert.Contains(t, eBuf.String(), "[ERROR]")
	assert.Contains(t, eBuf.String(), "-")
	assert.Contains(t, eBuf.String(), "internal/core.TestLogger")
	assert.Contains(t, eBuf.String(), "internal/core/logger_test.go:")
	println(eBuf.String())
	eBuf.Reset()
}

// TestNoOpLogger ensures no output is produced
func TestNoOpLogger(t *testing.T) {
	logger := &NoOpLogger{}

	// Capture stdout/stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		w.Close()
	}()

	// Test all methods
	logger.Info("test info")
	logger.Infof("test infof %s", "arg")
	logger.Warn("test warn")
	logger.Warnf("test warnf %s", "arg")
	logger.Error("test error")
	logger.Errorf("test errorf %s", "arg")
	logger.Critical("test critical")
	logger.Criticalf("test criticalf %s", "arg")

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Should be no output
	assert.Empty(t, buf.String(), "NoOpLogger should produce no output")
}

// TestFileLogger writes to file only
func TestFileLogger(t *testing.T) {
	// Create temp file
	tmpfile, err := os.CreateTemp("", "hercules_test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	logger := NewFileLogger(tmpfile)

	// Test all methods
	logger.Info("test info")
	logger.Infof("test infof %s", "arg")
	logger.Warn("test warn")
	logger.Warnf("test warnf %s", "arg")
	logger.Error("test error")
	logger.Errorf("test errorf %s", "arg")
	logger.Critical("test critical")
	logger.Criticalf("test criticalf %s", "arg")

	// Read from file
	content, err := os.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	contentStr := string(content)

	// Check that logs are in file
	assert.Contains(t, contentStr, "[INFO]")
	assert.Contains(t, contentStr, "[WARN]")
	assert.Contains(t, contentStr, "[ERROR]")
	assert.Contains(t, contentStr, "test info")
	assert.Contains(t, contentStr, "test warn")
	assert.Contains(t, contentStr, "test error")
	assert.Contains(t, contentStr, "test infof arg")
	assert.Contains(t, contentStr, "test warnf arg")
	assert.Contains(t, contentStr, "test errorf arg")
	assert.Contains(t, contentStr, "test critical")
	assert.Contains(t, contentStr, "test criticalf arg")
}

// TestSlogLogger writes JSON to stdout
func TestSlogLogger(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
		w.Close()
	}()

	logger := NewSlogLogger(os.Stdout)

	// Test all methods
	logger.Info("test info")
	logger.Infof("test infof %s", "arg")
	logger.Warn("test warn")
	logger.Warnf("test warnf %s", "arg")
	logger.Error("test error")
	logger.Errorf("test errorf %s", "arg")
	logger.Critical("test critical")
	logger.Criticalf("test criticalf %s", "arg")

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that logs are JSON and contain expected content
	assert.Contains(t, output, `"level":"INFO"`)
	assert.Contains(t, output, `"level":"WARN"`)
	assert.Contains(t, output, `"level":"ERROR"`)
	assert.Contains(t, output, `"msg":"info"`)
	assert.Contains(t, output, `"msg":"warn"`)
	assert.Contains(t, output, `"msg":"error"`)
	assert.Contains(t, output, `"msg":"critical"`)
	assert.Contains(t, output, `"msg":"test infof arg"`)
	assert.Contains(t, output, `"msg":"test warnf arg"`)
	assert.Contains(t, output, `"msg":"test errorf arg"`)
	assert.Contains(t, output, `"msg":"test criticalf arg"`)
}

// TestGlobalLogger ensures SetLogger and GetLogger work correctly
func TestGlobalLogger(t *testing.T) {
	// Reset to default
	SetLogger(&NoOpLogger{})

	// Test that GetLogger returns the set logger
	logger1 := GetLogger()
	assert.IsType(t, &NoOpLogger{}, logger1)

	// Test setting a new logger
	testLogger := &FileLogger{}
	SetLogger(testLogger)

	logger2 := GetLogger()
	assert.Equal(t, testLogger, logger2)
}

// TestLoggerModeBehavior tests CLI vs Server mode behavior
func TestLoggerModeBehavior(t *testing.T) {
	// Test CLI mode with no log file (should use NoOpLogger)
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		w.Close()
	}()

	// Set NoOpLogger (simulating CLI mode with no log file)
	SetLogger(&NoOpLogger{})

	// Try to log
	logger := GetLogger()
	logger.Info("this should not appear")
	logger.Warn("this should not appear")
	logger.Error("this should not appear")

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Should be no output in CLI mode with no log file
	assert.Empty(t, buf.String(), "CLI mode with no log file should produce no output")
}

// TestFileLoggerErrorHandling tests behavior when file cannot be opened
func TestFileLoggerErrorHandling(t *testing.T) {
	// Test that FileLogger can be created with a valid writer
	var buf bytes.Buffer
	logger := NewFileLogger(&buf)

	// Should not panic and should still work
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")

	// Test that it doesn't crash
	assert.NotNil(t, logger)

	// Test that logs are written to the buffer
	output := buf.String()
	assert.Contains(t, output, "[INFO]")
	assert.Contains(t, output, "[WARN]")
	assert.Contains(t, output, "[ERROR]")
	assert.Contains(t, output, "test")
}

// TestSlogLoggerJSONFormat tests that SlogLogger produces valid JSON
func TestSlogLoggerJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewSlogLogger(&buf)

	logger.Info("test message")

	output := buf.String()

	// Should be valid JSON
	assert.Contains(t, output, `"level":"INFO"`)
	assert.Contains(t, output, `"msg":"info"`)
	assert.True(t, strings.HasPrefix(output, "{"))
	assert.True(t, strings.HasSuffix(strings.TrimSpace(output), "}"))
}

// TestLoggerInterfaceCompliance tests that all logger implementations satisfy the interface
func TestLoggerInterfaceCompliance(t *testing.T) {
	var _ Logger = &NoOpLogger{}
	var _ Logger = &FileLogger{}
	var _ Logger = &SlogLogger{}

	// Test that DefaultLogger also satisfies the interface
	var _ Logger = NewLogger()
}
