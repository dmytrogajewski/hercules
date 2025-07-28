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
		return c.buildEmptyResult(), nil
	}

	config := c.DefaultConfig()
	commentDetails := c.analyzeCommentPlacement(comments, functions, config)
	metrics := c.calculateMetrics(commentDetails, functions)

	return c.buildResult(commentDetails, functions, metrics), nil
}

// FormatReport formats comment analysis results as human-readable text
func (c *CommentsAnalyzer) FormatReport(report analyze.Report, w io.Writer) error {
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
	commentBlocks := c.groupCommentsIntoBlocks(comments)
	return c.analyzeCommentBlocks(commentBlocks, functions, config)
}

// groupCommentsIntoBlocks groups consecutive comment lines into blocks
func (c *CommentsAnalyzer) groupCommentsIntoBlocks(comments []*node.Node) []CommentBlock {
	if len(comments) == 0 {
		return []CommentBlock{}
	}

	sortedComments := c.sortCommentsByLine(comments)
	return c.createCommentBlocks(sortedComments)
}

// sortCommentsByLine sorts comments by line number
func (c *CommentsAnalyzer) sortCommentsByLine(comments []*node.Node) []*node.Node {
	sortedComments := make([]*node.Node, len(comments))
	copy(sortedComments, comments)
	sort.Slice(sortedComments, func(i, j int) bool {
		if sortedComments[i].Pos == nil || sortedComments[j].Pos == nil {
			return false
		}
		return sortedComments[i].Pos.StartLine < sortedComments[j].Pos.StartLine
	})
	return sortedComments
}

// createCommentBlocks creates comment blocks from sorted comments
func (c *CommentsAnalyzer) createCommentBlocks(sortedComments []*node.Node) []CommentBlock {
	var blocks []CommentBlock
	var currentBlock CommentBlock

	for _, comment := range sortedComments {
		if comment.Pos == nil {
			continue
		}

		commentStart := int(comment.Pos.StartLine)
		commentEnd := int(comment.Pos.EndLine)

		if c.shouldStartNewBlock(currentBlock, commentStart) {
			blocks = c.addBlockIfValid(blocks, currentBlock)
			currentBlock = c.createNewBlock(comment, commentStart, commentEnd)
		} else {
			currentBlock = c.extendCurrentBlock(currentBlock, comment, commentEnd)
		}
	}

	return c.addBlockIfValid(blocks, currentBlock)
}

// shouldStartNewBlock determines if a new comment block should be started
func (c *CommentsAnalyzer) shouldStartNewBlock(currentBlock CommentBlock, commentStart int) bool {
	return len(currentBlock.Comments) == 0 || commentStart > currentBlock.EndLine+1
}

// addBlockIfValid adds a block to the list if it contains comments
func (c *CommentsAnalyzer) addBlockIfValid(blocks []CommentBlock, block CommentBlock) []CommentBlock {
	if len(block.Comments) > 0 {
		blocks = append(blocks, block)
	}
	return blocks
}

// createNewBlock creates a new comment block
func (c *CommentsAnalyzer) createNewBlock(comment *node.Node, startLine, endLine int) CommentBlock {
	return CommentBlock{
		Comments:  []*node.Node{comment},
		StartLine: startLine,
		EndLine:   endLine,
		FullText:  comment.Token,
	}
}

// extendCurrentBlock extends the current block with a new comment
func (c *CommentsAnalyzer) extendCurrentBlock(block CommentBlock, comment *node.Node, endLine int) CommentBlock {
	block.Comments = append(block.Comments, comment)
	block.EndLine = endLine
	block.FullText += "\n" + comment.Token
	return block
}

// analyzeCommentBlocks analyzes multiple comment blocks
func (c *CommentsAnalyzer) analyzeCommentBlocks(blocks []CommentBlock, functions []*node.Node, config CommentConfig) []CommentDetail {
	var details []CommentDetail
	for _, block := range blocks {
		blockDetails := c.analyzeCommentBlock(block, functions, config)
		details = append(details, blockDetails...)
	}
	return details
}

// analyzeCommentBlock analyzes a comment block as a single unit
func (c *CommentsAnalyzer) analyzeCommentBlock(block CommentBlock, functions []*node.Node, config CommentConfig) []CommentDetail {
	blockNode := c.createBlockNode(block)
	blockDetail := c.analyzeSingleComment(blockNode, functions, config)
	return c.createCommentDetails(block, blockDetail)
}

// createBlockNode creates a virtual comment node representing the entire block
func (c *CommentsAnalyzer) createBlockNode(block CommentBlock) *node.Node {
	return &node.Node{
		Type:  node.UASTComment,
		Token: block.FullText,
		Pos: &node.Positions{
			StartLine: uint(block.StartLine),
			EndLine:   uint(block.EndLine),
		},
	}
}

// createCommentDetails creates comment details for all comments in a block
func (c *CommentsAnalyzer) createCommentDetails(block CommentBlock, blockDetail CommentDetail) []CommentDetail {
	var details []CommentDetail
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
	lineNumber := c.getCommentLineNumber(comment)
	detail := c.createCommentDetail(comment, lineNumber)
	target := c.findClosestTarget(comment, functions)

	if target != nil {
		c.analyzeCommentWithTarget(comment, target, config, &detail)
	} else {
		c.analyzeCommentWithoutTarget(&detail)
	}

	return detail
}

// getCommentLineNumber gets the line number of a comment
func (c *CommentsAnalyzer) getCommentLineNumber(comment *node.Node) int {
	if comment.Pos != nil {
		return int(comment.Pos.StartLine)
	}
	return 0
}

// createCommentDetail creates a basic comment detail
func (c *CommentsAnalyzer) createCommentDetail(comment *node.Node, lineNumber int) CommentDetail {
	return CommentDetail{
		Type:       string(comment.Type),
		Token:      comment.Token,
		Score:      0.0,
		IsGood:     false,
		LineNumber: lineNumber,
	}
}

// analyzeCommentWithTarget analyzes a comment that has a target
func (c *CommentsAnalyzer) analyzeCommentWithTarget(comment *node.Node, target *node.Node, config CommentConfig, detail *CommentDetail) {
	detail.TargetType = string(target.Type)
	detail.TargetName = c.extractTargetName(target)
	detail.Position = c.determinePosition(comment, target)

	if c.isCommentProperlyPlaced(comment, target) {
		detail.Score = config.RewardScore
		detail.IsGood = true
	} else {
		detail.Score = c.getPenaltyScore(target, config)
		detail.IsGood = false
	}
}

// analyzeCommentWithoutTarget analyzes a comment without a target
func (c *CommentsAnalyzer) analyzeCommentWithoutTarget(detail *CommentDetail) {
	detail.Score = -0.2
	detail.IsGood = false
	detail.Position = "unassociated"
}

// getPenaltyScore gets the penalty score for a target type
func (c *CommentsAnalyzer) getPenaltyScore(target *node.Node, config CommentConfig) float64 {
	if penalty, exists := config.PenaltyScores[string(target.Type)]; exists {
		return penalty
	}
	return -0.1
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

	if commentEndLine < targetLine {
		return targetLine - commentEndLine
	}

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

	if commentEndLine >= targetLine {
		return false
	}

	gap := targetLine - commentEndLine
	return c.isGapAcceptable(commentStartLine, commentEndLine, gap)
}

// isGapAcceptable checks if the gap between comment and target is acceptable
func (c *CommentsAnalyzer) isGapAcceptable(commentStartLine, commentEndLine, gap int) bool {
	if commentStartLine == commentEndLine {
		return gap <= 2
	}
	return gap <= 3
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

// buildEmptyResult creates an empty result when no comments are found
func (c *CommentsAnalyzer) buildEmptyResult() analyze.Report {
	return common.NewResultBuilder().BuildCustomEmptyResult(map[string]interface{}{
		"total_comments":       0,
		"good_comments":        0,
		"bad_comments":         0,
		"overall_score":        0.0,
		"total_functions":      0,
		"documented_functions": 0,
		"message":              "No comments found",
	})
}

// buildResult builds the complete analysis result
func (c *CommentsAnalyzer) buildResult(commentDetails []CommentDetail, functions []*node.Node, metrics CommentMetrics) analyze.Report {
	commentDetailsInterface := c.buildCommentDetailsInterface(commentDetails)
	detailedCommentsTable := c.buildDetailedCommentsTable(commentDetails)
	detailedFunctionsTable := c.buildDetailedFunctionsTable(functions, metrics)
	functionSummaryInterface := c.buildFunctionSummaryInterface(metrics)

	return analyze.Report{
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
}

// buildCommentDetailsInterface builds the comment details interface
func (c *CommentsAnalyzer) buildCommentDetailsInterface(commentDetails []CommentDetail) []map[string]interface{} {
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
	return commentDetailsInterface
}

// buildDetailedCommentsTable builds the detailed comments table for display
func (c *CommentsAnalyzer) buildDetailedCommentsTable(commentDetails []CommentDetail) []map[string]interface{} {
	detailedCommentsTable := make([]map[string]interface{}, 0, len(commentDetails))
	for _, detail := range commentDetails {
		assessment := c.getCommentAssessment(detail.IsGood)
		commentBody := c.truncateCommentBody(detail.Token)

		detailedCommentsTable = append(detailedCommentsTable, map[string]interface{}{
			"line":       detail.LineNumber,
			"comment":    commentBody,
			"placement":  detail.Position,
			"target":     detail.TargetName,
			"assessment": assessment,
		})
	}
	return detailedCommentsTable
}

// buildDetailedFunctionsTable builds the detailed functions table for display
func (c *CommentsAnalyzer) buildDetailedFunctionsTable(functions []*node.Node, metrics CommentMetrics) []map[string]interface{} {
	detailedFunctionsTable := make([]map[string]interface{}, 0, len(functions))
	for _, function := range functions {
		funcName := c.extractTargetName(function)
		funcInfo := metrics.FunctionSummary[funcName]

		assessment, commentType := c.getFunctionAssessment(funcInfo)
		funcType := c.getFunctionType(function)
		lineCount := c.getFunctionLineCount(function)

		detailedFunctionsTable = append(detailedFunctionsTable, map[string]interface{}{
			"function":   funcName,
			"type":       funcType,
			"lines":      lineCount,
			"comment":    commentType,
			"assessment": assessment,
		})
	}
	return detailedFunctionsTable
}

// buildFunctionSummaryInterface builds the function summary interface
func (c *CommentsAnalyzer) buildFunctionSummaryInterface(metrics CommentMetrics) map[string]any {
	functionSummaryInterface := make(map[string]any)
	for name, info := range metrics.FunctionSummary {
		functionSummaryInterface[name] = map[string]any{
			"name":         info.Name,
			"type":         info.Type,
			"has_comment":  info.HasComment,
			"comment_type": info.CommentType,
		}
	}
	return functionSummaryInterface
}

// getCommentAssessment gets the assessment string for a comment
func (c *CommentsAnalyzer) getCommentAssessment(isGood bool) string {
	if isGood {
		return "✅ OK"
	}
	return "❌ Not OK"
}

// truncateCommentBody truncates comment body if too long for table display
func (c *CommentsAnalyzer) truncateCommentBody(commentBody string) string {
	if len(commentBody) > 50 {
		return commentBody[:47] + "..."
	}
	return commentBody
}

// getFunctionAssessment gets the assessment and comment type for a function
func (c *CommentsAnalyzer) getFunctionAssessment(funcInfo FunctionInfo) (string, string) {
	if funcInfo.HasComment {
		return "✅ Well Documented", funcInfo.CommentType
	}
	return "❌ No Comment", "None"
}

// getFunctionType gets the function type
func (c *CommentsAnalyzer) getFunctionType(function *node.Node) string {
	funcType := string(function.Type)
	if funcType == "" {
		return "Unknown"
	}
	return funcType
}

// getFunctionLineCount gets the line count of a function
func (c *CommentsAnalyzer) getFunctionLineCount(function *node.Node) int {
	if function.Pos != nil {
		return int(function.Pos.EndLine - function.Pos.StartLine + 1)
	}
	return 0
}
