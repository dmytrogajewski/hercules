package complexity

import (
	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
	"github.com/dmytrogajewski/hercules/pkg/analyzers/common"
)

// ComplexityAggregator aggregates results from multiple complexity analyses
type ComplexityAggregator struct {
	*common.Aggregator
	detailedFunctions []map[string]interface{}
}

// NewComplexityAggregator creates a new ComplexityAggregator
func NewComplexityAggregator() *ComplexityAggregator {
	numericKeys := getNumericKeys()
	countKeys := getCountKeys()
	messageBuilder := buildComplexityMessage
	emptyResultBuilder := buildEmptyComplexityResult

	return &ComplexityAggregator{
		Aggregator: common.NewAggregatorWithCustomEmptyResult(
			"complexity",
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
func (ca *ComplexityAggregator) Aggregate(results map[string]analyze.Report) {
	ca.collectDetailedFunctions(results)
	ca.Aggregator.Aggregate(results)
}

// GetResult overrides the base GetResult method to include detailed functions
func (ca *ComplexityAggregator) GetResult() analyze.Report {
	result := ca.Aggregator.GetResult()
	ca.addDetailedFunctionsToResult(result)
	return result
}

// collectDetailedFunctions extracts detailed functions from all reports
func (ca *ComplexityAggregator) collectDetailedFunctions(results map[string]analyze.Report) {
	for _, report := range results {
		if report == nil {
			continue
		}
		ca.extractFunctionsFromReport(report)
	}
}

// extractFunctionsFromReport extracts functions from a single report
func (ca *ComplexityAggregator) extractFunctionsFromReport(report analyze.Report) {
	if functions, ok := report["functions"].([]map[string]interface{}); ok {
		ca.detailedFunctions = append(ca.detailedFunctions, functions...)
	}
}

// addDetailedFunctionsToResult adds detailed functions to the result
func (ca *ComplexityAggregator) addDetailedFunctionsToResult(result analyze.Report) {
	if len(ca.detailedFunctions) > 0 {
		result["functions"] = ca.detailedFunctions
	}
}

// getNumericKeys returns the numeric keys for complexity analysis
func getNumericKeys() []string {
	return []string{"average_complexity", "cognitive_complexity", "nesting_depth"}
}

// getCountKeys returns the count keys for complexity analysis
func getCountKeys() []string {
	return []string{"total_functions", "max_complexity", "total_complexity", "decision_points"}
}

// buildComplexityMessage creates a message based on the complexity score
func buildComplexityMessage(score float64) string {
	switch {
	case score <= 1.0:
		return "Excellent complexity - functions are simple and maintainable"
	case score <= 3.0:
		return "Good complexity - functions have reasonable complexity"
	case score <= 7.0:
		return "Fair complexity - some functions could be simplified"
	default:
		return "High complexity - functions are complex and should be refactored"
	}
}

// buildEmptyComplexityResult creates an empty result with default values
func buildEmptyComplexityResult() analyze.Report {
	return analyze.Report{
		"total_functions":    0,
		"average_complexity": 0.0,
		"max_complexity":     0,
		"total_complexity":   0,
		"message":            "No functions found",
	}
}
