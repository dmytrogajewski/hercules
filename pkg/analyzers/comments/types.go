package comments

import (
	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
	"github.com/dmytrogajewski/hercules/pkg/analyzers/common"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

// CommentsAnalyzer provides comment placement analysis
type CommentsAnalyzer struct {
	traverser *common.UASTTraverser
	extractor *common.DataExtractor
}

// CommentMetrics holds comment analysis results
type CommentMetrics struct {
	TotalComments       int                     `json:"total_comments"`
	GoodComments        int                     `json:"good_comments"`
	BadComments         int                     `json:"bad_comments"`
	OverallScore        float64                 `json:"overall_score"`
	CommentDetails      []CommentDetail         `json:"comment_details"`
	FunctionSummary     map[string]FunctionInfo `json:"function_summary"`
	TotalFunctions      int                     `json:"total_functions"`
	DocumentedFunctions int                     `json:"documented_functions"`
}

// CommentDetail holds information about a specific comment
type CommentDetail struct {
	Type           string  `json:"type"`
	Token          string  `json:"token"`
	Position       string  `json:"position"`
	Score          float64 `json:"score"`
	IsGood         bool    `json:"is_good"`
	TargetType     string  `json:"target_type"`
	TargetName     string  `json:"target_name"`
	LineNumber     int     `json:"line_number"`
	StartLine      int     `json:"start_line"`
	EndLine        int     `json:"end_line"`
	Length         int     `json:"length"`
	Quality        string  `json:"quality"`
	Recommendation string  `json:"recommendation"`
}

// FunctionInfo holds information about a function
type FunctionInfo struct {
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	HasComment    bool    `json:"has_comment"`
	CommentType   string  `json:"comment_type"`
	StartLine     int     `json:"start_line"`
	EndLine       int     `json:"end_line"`
	CommentScore  float64 `json:"comment_score"`
	Documentation string  `json:"documentation"`
	NeedsComment  bool    `json:"needs_comment"`
}

// CommentConfig holds configuration for comment analysis
type CommentConfig struct {
	RewardScore      float64
	PenaltyScores    map[string]float64
	MaxCommentLength int
}

// CommentBlock represents a group of consecutive comment lines
type CommentBlock struct {
	Comments  []*node.Node
	StartLine int
	EndLine   int
	FullText  string
}

// NewCommentsAnalyzer creates a new CommentsAnalyzer with generic components
func NewCommentsAnalyzer() *CommentsAnalyzer {
	// Configure UAST traverser for comments and functions
	traversalConfig := common.TraversalConfig{
		Filters: []common.NodeFilter{
			{
				Types: []string{node.UASTComment},
				Roles: []string{node.RoleComment},
			},
			{
				Types: []string{node.UASTFunction, node.UASTMethod, node.UASTClass, node.UASTInterface, node.UASTStruct},
				Roles: []string{node.RoleFunction, node.RoleDeclaration},
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
			"comment_text": func(n *node.Node) (string, bool) {
				if n == nil || n.Token == "" {
					return "", false
				}
				return n.Token, true
			},
		},
	}

	return &CommentsAnalyzer{
		traverser: common.NewUASTTraverser(traversalConfig),
		extractor: common.NewDataExtractor(extractionConfig),
	}
}

// CreateAggregator creates a new aggregator for comment analysis
func (c *CommentsAnalyzer) CreateAggregator() analyze.ResultAggregator {
	return NewCommentsAggregator()
}

// DefaultConfig returns the default configuration for comment analysis
func (c *CommentsAnalyzer) DefaultConfig() CommentConfig {
	return CommentConfig{
		RewardScore: 1.0,
		PenaltyScores: map[string]float64{
			node.UASTFunction:   -0.5,
			node.UASTMethod:     -0.5,
			node.UASTClass:      -0.3,
			node.UASTInterface:  -0.3,
			node.UASTStruct:     -0.3,
			node.UASTVariable:   -0.1,
			node.UASTAssignment: -0.1,
			node.UASTCall:       -0.1,
		},
		MaxCommentLength: 500,
	}
}
