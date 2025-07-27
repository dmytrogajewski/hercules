package halstead

import (
	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
	"github.com/dmytrogajewski/hercules/pkg/analyzers/common"
)

// HalsteadAggregator aggregates Halstead analysis results
type HalsteadAggregator struct {
	*common.Aggregator
	detailedFunctions []map[string]interface{}
}

// NewHalsteadAggregator creates a new Halstead aggregator
func NewHalsteadAggregator() *HalsteadAggregator {
	numericKeys := getNumericKeys()
	countKeys := getCountKeys()
	messageBuilder := buildHalsteadMessage
	emptyResultBuilder := buildEmptyHalsteadResult

	return &HalsteadAggregator{
		Aggregator: common.NewAggregatorWithCustomEmptyResult(
			"halstead",
			numericKeys,
			countKeys,
			"functions",
			"name",
			messageBuilder,
			emptyResultBuilder,
		),
		detailedFunctions: make([]map[string]interface{}, 0),
	}
}

// Aggregate overrides the base Aggregate method to collect detailed functions
func (ha *HalsteadAggregator) Aggregate(results map[string]analyze.Report) {
	ha.collectDetailedFunctions(results)
	ha.Aggregator.Aggregate(results)
}

// GetResult overrides the base GetResult method to include detailed functions
func (ha *HalsteadAggregator) GetResult() analyze.Report {
	result := ha.Aggregator.GetResult()
	ha.addDetailedFunctionsToResult(result)
	return result
}

// collectDetailedFunctions extracts detailed functions from all reports
func (ha *HalsteadAggregator) collectDetailedFunctions(results map[string]analyze.Report) {
	for _, report := range results {
		if report == nil {
			continue
		}
		ha.extractFunctionsFromReport(report)
	}
}

// extractFunctionsFromReport extracts functions from a single report
func (ha *HalsteadAggregator) extractFunctionsFromReport(report analyze.Report) {
	if functions, ok := report["functions"].([]map[string]interface{}); ok {
		ha.detailedFunctions = append(ha.detailedFunctions, functions...)
	}
}

// addDetailedFunctionsToResult adds detailed functions to the result
func (ha *HalsteadAggregator) addDetailedFunctionsToResult(result analyze.Report) {
	if len(ha.detailedFunctions) > 0 {
		result["functions"] = ha.detailedFunctions
	}
}

// getNumericKeys returns the numeric keys for Halstead aggregation
func getNumericKeys() []string {
	return []string{
		"volume",
		"difficulty",
		"effort",
		"time_to_program",
		"delivered_bugs",
		"distinct_operators",
		"distinct_operands",
		"total_operators",
		"total_operands",
		"vocabulary",
		"length",
		"estimated_length",
	}
}

// getCountKeys returns the count keys for Halstead aggregation
func getCountKeys() []string {
	return []string{"total_functions"}
}

// buildHalsteadMessage creates a message based on the volume metric
func buildHalsteadMessage(volume float64) string {
	switch {
	case volume >= 5000:
		return "Very high Halstead complexity - significant refactoring recommended"
	case volume >= 1000:
		return "High Halstead complexity - consider refactoring"
	case volume >= 100:
		return "Moderate Halstead complexity - acceptable"
	default:
		return "Low Halstead complexity - well-structured code"
	}
}

// buildEmptyHalsteadResult creates an empty result with default Halstead values
func buildEmptyHalsteadResult() analyze.Report {
	return analyze.Report{
		"total_functions":    0,
		"volume":             0.0,
		"difficulty":         0.0,
		"effort":             0.0,
		"time_to_program":    0.0,
		"delivered_bugs":     0.0,
		"distinct_operators": 0,
		"distinct_operands":  0,
		"total_operators":    0,
		"total_operands":     0,
		"vocabulary":         0,
		"length":             0,
		"estimated_length":   0.0,
		"message":            "No functions found",
	}
}
