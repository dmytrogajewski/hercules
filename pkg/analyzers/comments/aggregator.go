package comments

import (
	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
	"github.com/dmytrogajewski/hercules/pkg/analyzers/common"
)

// CommentsAggregator aggregates results from multiple comment analyses
type CommentsAggregator struct {
	*common.Aggregator
	detailedComments  []map[string]interface{}
	detailedFunctions []map[string]interface{}
}

// NewCommentsAggregator creates a new CommentsAggregator
func NewCommentsAggregator() *CommentsAggregator {
	numericKeys := []string{"overall_score", "good_comments_ratio", "documentation_coverage"}
	countKeys := []string{"total_comments", "good_comments", "bad_comments", "total_functions", "documented_functions", "total_comment_details"}

	messageBuilder := func(score float64) string {
		if score >= 0.8 {
			return "Excellent comment quality and placement"
		}
		if score >= 0.6 {
			return "Good comment quality with room for improvement"
		}
		if score >= 0.4 {
			return "Fair comment quality - consider improving placement"
		}
		return "Poor comment quality - significant improvement needed"
	}

	emptyResultBuilder := func() analyze.Report {
		return analyze.Report{
			"total_comments":        0,
			"good_comments":         0,
			"bad_comments":          0,
			"overall_score":         0.0,
			"total_functions":       0,
			"documented_functions":  0,
			"total_comment_details": 0,
			"message":               "No comments found",
		}
	}

	return &CommentsAggregator{
		Aggregator: common.NewAggregatorWithCustomEmptyResult(
			"comments",
			numericKeys,
			countKeys,
			"comments",
			"line",
			messageBuilder,
			emptyResultBuilder,
		),
		detailedComments:  make([]map[string]interface{}, 0),
		detailedFunctions: make([]map[string]interface{}, 0),
	}
}

// Aggregate overrides the base Aggregate method to collect detailed comments and functions
func (ca *CommentsAggregator) Aggregate(results map[string]analyze.Report) {
	// Collect detailed comments and functions from all reports first
	for _, report := range results {
		if report == nil {
			continue
		}

		if comments, ok := report["comments"].([]map[string]interface{}); ok {
			ca.detailedComments = append(ca.detailedComments, comments...)
		}

		if functions, ok := report["functions"].([]map[string]interface{}); ok {
			ca.detailedFunctions = append(ca.detailedFunctions, functions...)
		}
	}

	// Call the base aggregate method
	ca.Aggregator.Aggregate(results)
}

// GetResult overrides the base GetResult method to include detailed comments and functions
func (ca *CommentsAggregator) GetResult() analyze.Report {
	// Get the base result
	result := ca.Aggregator.GetResult()

	// Add the detailed comments table if we have any
	if len(ca.detailedComments) > 0 {
		result["comments"] = ca.detailedComments
	}

	// Add the detailed functions table if we have any
	if len(ca.detailedFunctions) > 0 {
		result["functions"] = ca.detailedFunctions
	}

	return result
}
