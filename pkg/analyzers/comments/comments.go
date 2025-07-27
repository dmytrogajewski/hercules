package comments

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
	"github.com/dmytrogajewski/hercules/pkg/analyzers/common"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

// Name returns the analyzer name
func (c *CommentsAnalyzer) Name() string {
	return "comments"
}

// Thresholds returns the quality thresholds for this analyzer
func (c *CommentsAnalyzer) Thresholds() analyze.Thresholds {
	return analyze.Thresholds{
		"overall_score": {
			"red":    0.4,
			"yellow": 0.6,
			"green":  0.8,
		},
		"good_comments_ratio": {
			"red":    0.4,
			"yellow": 0.6,
			"green":  0.8,
		},
		"documentation_coverage": {
			"red":    0.4,
			"yellow": 0.6,
			"green":  0.8,
		},
	}
}

// Analyze performs comment analysis using default configuration
func (c *CommentsAnalyzer) Analyze(root *node.Node) (analyze.Report, error) {
	if root == nil {
		return nil, fmt.Errorf("root node is nil")
	}

	comments := c.findComments(root)
	functions := c.findFunctions(root)

	if len(comments) == 0 {
		return common.NewResultBuilder().BuildCustomEmptyResult(map[string]interface{}{
			"total_comments":       0,
			"good_comments":        0,
			"bad_comments":         0,
			"overall_score":        0.0,
			"total_functions":      0,
			"documented_functions": 0,
			"message":              "No comments found",
		}), nil
	}

	config := c.DefaultConfig()
	commentDetails := c.analyzeCommentPlacement(comments, functions, config)
	metrics := c.calculateMetrics(commentDetails, functions)

	// Build comment details for result
	commentDetailsInterface := make([]map[string]interface{}, 0, len(commentDetails))
	for _, detail := range commentDetails {
		commentDetailsInterface = append(commentDetailsInterface, map[string]interface{}{
			"type":        detail.Type,
			"token":       detail.Token,
			"position":    detail.Position,
			"score":       detail.Score,
			"is_good":     detail.IsGood,
			"target_type": detail.TargetType,
			"target_name": detail.TargetName,
			"line_number": detail.LineNumber,
		})
	}

	// Build detailed comments table for display
	detailedCommentsTable := make([]map[string]interface{}, 0, len(commentDetails))
	for _, detail := range commentDetails {
		assessment := "❌ Not OK"
		if detail.IsGood {
			assessment = "✅ OK"
		}

		// Truncate comment body if too long for table display
		commentBody := detail.Token
		if len(commentBody) > 50 {
			commentBody = commentBody[:47] + "..."
		}

		detailedCommentsTable = append(detailedCommentsTable, map[string]interface{}{
			"line":       detail.LineNumber,
			"comment":    commentBody,
			"placement":  detail.Position,
			"target":     detail.TargetName,
			"assessment": assessment,
		})
	}

	// Build detailed functions table for display
	detailedFunctionsTable := make([]map[string]interface{}, 0, len(functions))
	for _, function := range functions {
		funcName := c.extractTargetName(function)
		funcInfo := metrics.FunctionSummary[funcName]

		// Determine assessment based on documentation status
		assessment := "❌ No Comment"
		commentType := "None"

		if funcInfo.HasComment {
			assessment = "✅ Well Documented"
			commentType = funcInfo.CommentType
		}

		// Get function type
		funcType := string(function.Type)
		if funcType == "" {
			funcType = "Unknown"
		}

		// Get line count
		lineCount := 0
		if function.Pos != nil {
			lineCount = int(function.Pos.EndLine - function.Pos.StartLine + 1)
		}

		detailedFunctionsTable = append(detailedFunctionsTable, map[string]interface{}{
			"function":   funcName,
			"type":       funcType,
			"lines":      lineCount,
			"comment":    commentType,
			"assessment": assessment,
		})
	}

	// Build function summary for result
	functionSummaryInterface := make(map[string]interface{})
	for name, info := range metrics.FunctionSummary {
		functionSummaryInterface[name] = map[string]interface{}{
			"name":         info.Name,
			"type":         info.Type,
			"has_comment":  info.HasComment,
			"comment_type": info.CommentType,
		}
	}

	// Build the result with the expected structure
	result := analyze.Report{
		"total_comments":         metrics.TotalComments,
		"good_comments":          metrics.GoodComments,
		"bad_comments":           metrics.BadComments,
		"overall_score":          metrics.OverallScore,
		"total_functions":        metrics.TotalFunctions,
		"documented_functions":   metrics.DocumentedFunctions,
		"good_comments_ratio":    float64(metrics.GoodComments) / float64(metrics.TotalComments),
		"documentation_coverage": float64(metrics.DocumentedFunctions) / float64(metrics.TotalFunctions),
		"total_comment_details":  len(commentDetails),
		"comment_details":        commentDetailsInterface,
		"comments":               detailedCommentsTable,
		"functions":              detailedFunctionsTable,
		"function_summary":       functionSummaryInterface,
		"message":                c.getCommentMessage(metrics.OverallScore),
	}

	return result, nil
}

// FormatReport formats comment analysis results as human-readable text
func (c *CommentsAnalyzer) FormatReport(report analyze.Report, w io.Writer) error {
	// Use the generic formatter for the rest with explicit table display
	formatter := common.NewFormatter(common.FormatConfig{
		ShowProgressBars: true,
		ShowTables:       true,
		ShowDetails:      true,
		SkipHeader:       true,
	})
	formatted := formatter.FormatReport(report)
	_, err := fmt.Fprint(w, formatted)
	return err
}

// FormatReportJSON formats comment analysis results as JSON
func (c *CommentsAnalyzer) FormatReportJSON(report analyze.Report, w io.Writer) error {
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, string(jsonData))
	return err
}

// findComments finds all comment nodes using the generic traverser
func (c *CommentsAnalyzer) findComments(root *node.Node) []*node.Node {
	return c.traverser.FindNodesByType(root, []string{node.UASTComment})
}

// findFunctions finds all function/method/class nodes using the generic traverser
func (c *CommentsAnalyzer) findFunctions(root *node.Node) []*node.Node {
	functionTypes := []string{
		node.UASTFunction,
		node.UASTMethod,
		node.UASTClass,
		node.UASTInterface,
		node.UASTStruct,
	}
	return c.traverser.FindNodesByType(root, functionTypes)
}

// analyzeCommentPlacement analyzes the placement of comments relative to their targets
func (c *CommentsAnalyzer) analyzeCommentPlacement(comments []*node.Node, functions []*node.Node, config CommentConfig) []CommentDetail {
	var details []CommentDetail

	// First, group comments into blocks
	commentBlocks := c.groupCommentsIntoBlocks(comments)

	// Then analyze each block as a unit
	for _, block := range commentBlocks {
		blockDetails := c.analyzeCommentBlock(block, functions, config)
		details = append(details, blockDetails...)
	}

	return details
}

// groupCommentsIntoBlocks groups consecutive comment lines into blocks
func (c *CommentsAnalyzer) groupCommentsIntoBlocks(comments []*node.Node) []CommentBlock {
	if len(comments) == 0 {
		return []CommentBlock{}
	}

	var blocks []CommentBlock
	var currentBlock CommentBlock

	// Sort comments by line number
	sortedComments := make([]*node.Node, len(comments))
	copy(sortedComments, comments)
	sort.Slice(sortedComments, func(i, j int) bool {
		if sortedComments[i].Pos == nil || sortedComments[j].Pos == nil {
			return false
		}
		return sortedComments[i].Pos.StartLine < sortedComments[j].Pos.StartLine
	})

	for _, comment := range sortedComments {
		if comment.Pos == nil {
			continue
		}

		commentStart := int(comment.Pos.StartLine)
		commentEnd := int(comment.Pos.EndLine)

		// If this is the first comment or if there's a gap > 1 line, start a new block
		if len(currentBlock.Comments) == 0 || commentStart > currentBlock.EndLine+1 {
			// Save previous block if it exists
			if len(currentBlock.Comments) > 0 {
				blocks = append(blocks, currentBlock)
			}

			// Start new block
			currentBlock = CommentBlock{
				Comments:  []*node.Node{comment},
				StartLine: commentStart,
				EndLine:   commentEnd,
				FullText:  comment.Token,
			}
		} else {
			// Add to current block
			currentBlock.Comments = append(currentBlock.Comments, comment)
			currentBlock.EndLine = commentEnd
			currentBlock.FullText += "\n" + comment.Token
		}
	}

	// Add the last block
	if len(currentBlock.Comments) > 0 {
		blocks = append(blocks, currentBlock)
	}

	return blocks
}

// analyzeCommentBlock analyzes a comment block as a single unit
func (c *CommentsAnalyzer) analyzeCommentBlock(block CommentBlock, functions []*node.Node, config CommentConfig) []CommentDetail {
	var details []CommentDetail

	// Create a virtual comment node representing the entire block
	blockNode := &node.Node{
		Type:  node.UASTComment,
		Token: block.FullText,
		Pos: &node.Positions{
			StartLine: uint(block.StartLine),
			EndLine:   uint(block.EndLine),
		},
	}

	// Analyze the block as a single comment
	blockDetail := c.analyzeSingleComment(blockNode, functions, config)

	// If the block is good, mark all individual comments as good
	// If the block is bad, mark all individual comments as bad
	for _, comment := range block.Comments {
		detail := CommentDetail{
			Type:       string(comment.Type),
			Token:      comment.Token,
			Score:      blockDetail.Score,
			IsGood:     blockDetail.IsGood,
			TargetType: blockDetail.TargetType,
			TargetName: blockDetail.TargetName,
			Position:   blockDetail.Position,
			LineNumber: int(comment.Pos.StartLine),
		}
		details = append(details, detail)
	}

	return details
}

// analyzeSingleComment analyzes a single comment's placement and quality
func (c *CommentsAnalyzer) analyzeSingleComment(comment *node.Node, functions []*node.Node, config CommentConfig) CommentDetail {
	lineNumber := 0
	if comment.Pos != nil {
		lineNumber = int(comment.Pos.StartLine)
	}

	detail := CommentDetail{
		Type:       string(comment.Type),
		Token:      comment.Token,
		Score:      0.0,
		IsGood:     false,
		LineNumber: lineNumber,
	}

	target := c.findClosestTarget(comment, functions)

	if target != nil {
		detail.TargetType = string(target.Type)
		detail.TargetName = c.extractTargetName(target)
		detail.Position = c.determinePosition(comment, target)

		if c.isCommentProperlyPlaced(comment, target) {
			detail.Score = config.RewardScore
			detail.IsGood = true
		} else {
			if penalty, exists := config.PenaltyScores[string(target.Type)]; exists {
				detail.Score = penalty
			} else {
				detail.Score = -0.1
			}
			detail.IsGood = false
		}
	} else {
		detail.Score = -0.2
		detail.IsGood = false
		detail.Position = "unassociated"
	}

	return detail
}

// findClosestTarget finds the closest function/class to a comment
func (c *CommentsAnalyzer) findClosestTarget(comment *node.Node, functions []*node.Node) *node.Node {
	var closest *node.Node
	var minDistance int = -1

	for _, function := range functions {
		distance := c.calculateDistance(comment, function)
		if minDistance == -1 || distance < minDistance {
			minDistance = distance
			closest = function
		}
	}

	return closest
}

// calculateDistance calculates the line distance between comment and target
func (c *CommentsAnalyzer) calculateDistance(comment *node.Node, target *node.Node) int {
	if comment.Pos == nil || target.Pos == nil {
		return 999
	}

	commentEndLine := int(comment.Pos.EndLine)
	targetLine := int(target.Pos.StartLine)

	// If comment is above target, use distance from comment end to target start
	if commentEndLine < targetLine {
		return targetLine - commentEndLine
	}

	// If comment is below target, use a much higher penalty
	// This encourages associating comments with functions that come after them
	return 1000 + (commentEndLine - targetLine)
}

// isCommentProperlyPlaced checks if a comment is properly placed above its target
func (c *CommentsAnalyzer) isCommentProperlyPlaced(comment *node.Node, target *node.Node) bool {
	if comment.Pos == nil || target.Pos == nil {
		return false
	}

	commentStartLine := int(comment.Pos.StartLine)
	commentEndLine := int(comment.Pos.EndLine)
	targetLine := int(target.Pos.StartLine)

	// Check if comment is above the target (comment ends before target starts)
	if commentEndLine >= targetLine {
		return false
	}

	// Calculate the gap between comment and target
	gap := targetLine - commentEndLine

	// For single-line comments, allow up to 2 lines gap
	// For multi-line comments, allow up to 3 lines gap
	if commentStartLine == commentEndLine {
		return gap <= 2
	} else {
		return gap <= 3
	}
}

// determinePosition determines the relative position of comment to target
func (c *CommentsAnalyzer) determinePosition(comment *node.Node, target *node.Node) string {
	if comment.Pos == nil || target.Pos == nil {
		return "unknown"
	}

	commentEndLine := int(comment.Pos.EndLine)
	targetLine := int(target.Pos.StartLine)

	if commentEndLine < targetLine {
		return "above"
	} else if int(comment.Pos.StartLine) > targetLine {
		return "below"
	} else {
		return "inline"
	}
}

// extractTargetName extracts the name of a target node using generic extractor
func (c *CommentsAnalyzer) extractTargetName(target *node.Node) string {
	if name, ok := c.extractor.ExtractName(target, "function_name"); ok && name != "" {
		return name
	}
	if name, ok := common.ExtractFunctionName(target); ok && name != "" {
		return name
	}
	return "unknown"
}

// calculateMetrics calculates overall metrics from comment details and functions
func (c *CommentsAnalyzer) calculateMetrics(details []CommentDetail, functions []*node.Node) CommentMetrics {
	metrics := CommentMetrics{
		TotalComments:       len(details),
		GoodComments:        0,
		BadComments:         0,
		OverallScore:        0.0,
		CommentDetails:      details,
		FunctionSummary:     make(map[string]FunctionInfo),
		TotalFunctions:      len(functions),
		DocumentedFunctions: 0,
	}

	c.countCommentQuality(details, &metrics)
	c.calculateOverallScore(&metrics)
	c.buildFunctionSummary(functions, details, &metrics)

	return metrics
}

// countCommentQuality counts good and bad comments
func (c *CommentsAnalyzer) countCommentQuality(details []CommentDetail, metrics *CommentMetrics) {
	for _, detail := range details {
		if detail.IsGood {
			metrics.GoodComments++
		} else {
			metrics.BadComments++
		}
	}
}

// calculateOverallScore calculates the overall comment quality score
func (c *CommentsAnalyzer) calculateOverallScore(metrics *CommentMetrics) {
	if metrics.TotalComments > 0 {
		metrics.OverallScore = float64(metrics.GoodComments) / float64(metrics.TotalComments)
	}
}

// buildFunctionSummary builds the function summary with documentation status
func (c *CommentsAnalyzer) buildFunctionSummary(functions []*node.Node, details []CommentDetail, metrics *CommentMetrics) {
	for _, function := range functions {
		funcName := c.extractTargetName(function)
		funcInfo := FunctionInfo{
			Name:       funcName,
			Type:       string(function.Type),
			HasComment: false,
		}

		if c.hasGoodComment(funcName, details) {
			funcInfo.HasComment = true
			funcInfo.CommentType = c.getCommentType(funcName, details)
			metrics.DocumentedFunctions++
		}

		metrics.FunctionSummary[funcName] = funcInfo
	}
}

// hasGoodComment checks if a function has a good comment
func (c *CommentsAnalyzer) hasGoodComment(funcName string, details []CommentDetail) bool {
	for _, detail := range details {
		if detail.TargetName == funcName && detail.IsGood {
			return true
		}
	}
	return false
}

// getCommentType gets the comment type for a function
func (c *CommentsAnalyzer) getCommentType(funcName string, details []CommentDetail) string {
	for _, detail := range details {
		if detail.TargetName == funcName && detail.IsGood {
			return detail.Type
		}
	}
	return ""
}

// getCommentMessage returns a message based on the comment quality score
func (c *CommentsAnalyzer) getCommentMessage(score float64) string {
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
