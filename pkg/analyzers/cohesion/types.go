package cohesion

import (
	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
	"github.com/dmytrogajewski/hercules/pkg/analyzers/common"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

// CohesionAnalyzer performs cohesion analysis on UAST
type CohesionAnalyzer struct {
	traverser  *common.UASTTraverser
	extractor  *common.DataExtractor
	aggregator *common.Aggregator
}

// Function represents a function with its cohesion metrics
type Function struct {
	Name      string
	LineCount int
	Variables []string
	Cohesion  float64
}

// NewCohesionAnalyzer creates a new CohesionAnalyzer with generic components
func NewCohesionAnalyzer() *CohesionAnalyzer {
	// Configure UAST traverser for functions
	traversalConfig := common.TraversalConfig{
		Filters: []common.NodeFilter{
			{
				Types: []string{node.UASTFunction, node.UASTMethod},
				Roles: []string{node.RoleFunction},
			},
		},
		MaxDepth: 10,
	}

	// Configure data extractor
	extractionConfig := common.ExtractionConfig{
		DefaultExtractors: true,
		NameExtractors: map[string]common.NameExtractor{
			"function_name": func(n *node.Node) (string, bool) {
				return common.ExtractFunctionName(n)
			},
			"variable_name": func(n *node.Node) (string, bool) {
				return common.ExtractVariableName(n)
			},
		},
	}

	// Configure aggregator
	numericKeys := []string{"lcom", "cohesion_score", "function_cohesion"}
	countKeys := []string{"total_functions"}

	messageBuilder := func(score float64) string {
		if score >= 0.8 {
			return "Excellent overall cohesion across all analyzed code"
		}
		if score >= 0.6 {
			return "Good overall cohesion with room for improvement"
		}
		if score >= 0.3 {
			return "Fair overall cohesion - consider refactoring some functions"
		}
		return "Poor overall cohesion - significant refactoring recommended"
	}

	emptyResultBuilder := func() analyze.Report {
		return analyze.Report{
			"total_functions":   0,
			"lcom":              0.0,
			"cohesion_score":    1.0,
			"function_cohesion": 1.0,
			"message":           "No functions found",
		}
	}

	return &CohesionAnalyzer{
		traverser: common.NewUASTTraverser(traversalConfig),
		extractor: common.NewDataExtractor(extractionConfig),
		aggregator: common.NewAggregatorWithCustomEmptyResult(
			"cohesion",
			numericKeys,
			countKeys,
			"functions",
			"name",
			messageBuilder,
			emptyResultBuilder,
		),
	}
}

// CreateAggregator creates a new aggregator for cohesion analysis
func (c *CohesionAnalyzer) CreateAggregator() analyze.ResultAggregator {
	return &CohesionAggregator{
		Aggregator: c.aggregator,
	}
}
