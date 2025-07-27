package common

import (
	"strings"
	"testing"

	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
)

func TestFormatter_CompactTableFormatting(t *testing.T) {
	config := FormatConfig{
		ShowTables:  true,
		ShowDetails: false,
		MaxItems:    10,
	}

	formatter := NewFormatter(config)

	// Create a test report with collection data
	report := analyze.Report{
		"analyzer_name": "test_analyzer",
		"message":       "Test analysis completed",
		"test_data": []map[string]interface{}{
			{"name": "item1", "value": 10, "score": 0.85},
			{"name": "item2", "value": 20, "score": 0.92},
			{"name": "item3", "value": 15, "score": 0.78},
		},
	}

	result := formatter.FormatReport(report)

	// Verify the result contains the table
	if !strings.Contains(result, "test_data:") {
		t.Error("Expected table to be included in formatted report")
	}

	// Verify the table is compact (no borders, minimal spacing)
	if strings.Contains(result, "│") {
		t.Error("Expected compact table without border characters")
	}

	// Verify data is present
	if !strings.Contains(result, "item1") || !strings.Contains(result, "item2") || !strings.Contains(result, "item3") {
		t.Error("Expected table to contain test data items")
	}

	// Verify footer is present (check for any footer text)
	if !strings.Contains(result, "TOTAL:") {
		t.Error("Expected table footer with item count")
	}
}
