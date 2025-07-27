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
	numericKeys := []string{"average_complexity", "cognitive_complexity", "nesting_depth"}
	countKeys := []string{"total_functions", "max_complexity", "total_complexity", "decision_points"}

	messageBuilder := func(score float64) string {
		if score <= 1.0 {
			return "Excellent complexity - functions are simple and maintainable"
		}
		if score <= 3.0 {
			return "Good complexity - functions have reasonable complexity"
		}
		if score <= 7.0 {
			return "Fair complexity - some functions could be simplified"
		}
		return "High complexity - functions are complex and should be refactored"
	}

	emptyResultBuilder := func() analyze.Report {
		return analyze.Report{
			"total_functions":    0,
			"average_complexity": 0.0,
			"max_complexity":     0,
			"total_complexity":   0,
			"message":            "No functions found",
		}
	}

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
	// Collect detailed functions from all reports first
	for _, report := range results {
		if report == nil {
			continue
		}
		if functions, ok := report["functions"].([]map[string]interface{}); ok {
			ca.detailedFunctions = append(ca.detailedFunctions, functions...)
		}
	}

	// Call the base aggregate method
	ca.Aggregator.Aggregate(results)
}

// GetResult overrides the base GetResult method to include detailed functions
func (ca *ComplexityAggregator) GetResult() analyze.Report {
	// Get the base result
	result := ca.Aggregator.GetResult()

	// Add the detailed functions table if we have any
	if len(ca.detailedFunctions) > 0 {
		result["functions"] = ca.detailedFunctions
	}

	return result
}
