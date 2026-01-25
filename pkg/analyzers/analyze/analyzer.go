package analyze

import (
	"fmt"
	"io"

	"github.com/dmytrogajewski/hercules/pkg/pipeline"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

type Report = map[string]any

// Thresholds represents color-coded thresholds for multiple metrics
// Structure: {"metric_name": {"red": value, "yellow": value, "green": value}}
type Thresholds = map[string]map[string]any

// Analyzer is the common base interface for all analyzers
type Analyzer interface {
	Name() string
	Flag() string
	Description() string

	// Configuration
	ListConfigurationOptions() []pipeline.ConfigurationOption
	Configure(facts map[string]interface{}) error
}

// StaticAnalyzer interface defines the contract for UAST-based static analysis
type StaticAnalyzer interface {
	Analyzer

	// Core analysis methods
	Analyze(root *node.Node) (Report, error)
	Thresholds() Thresholds

	// Aggregation methods
	CreateAggregator() ResultAggregator

	// Formatting methods
	FormatReport(report Report, writer io.Writer) error
	FormatReportJSON(report Report, writer io.Writer) error
}

// VisitorProvider enables single-pass traversal optimization
type VisitorProvider interface {
	CreateVisitor() AnalysisVisitor
}

// ResultAggregator defines the interface for aggregating analyzer results
type ResultAggregator interface {
	Aggregate(results map[string]Report)
	GetResult() Report
}

type Factory struct {
	analyzers map[string]StaticAnalyzer
}

func (f *Factory) RegisterAnalyzer(analyzer StaticAnalyzer) {
	f.analyzers[analyzer.Name()] = analyzer
}

func (f *Factory) RunAnalyzer(name string, root *node.Node) (Report, error) {
	analyzer, ok := f.analyzers[name]

	if !ok {
		return nil, fmt.Errorf("no registered analyzer with name=%s", name)
	}

	return analyzer.Analyze(root)
}

func (f *Factory) RunAnalyzers(root *node.Node, analyzers []string) (map[string]Report, error) {
	combinedReport := map[string]Report{}

	visitors := make([]NodeVisitor, 0)
	visitorAnalyzers := make(map[string]AnalysisVisitor)
	legacyAnalyzers := make([]string, 0)

	for _, name := range analyzers {
		analyzer, ok := f.analyzers[name]
		if !ok {
			return nil, fmt.Errorf("no registered analyzer with name=%s", name)
		}

		if vp, ok := analyzer.(VisitorProvider); ok {
			v := vp.CreateVisitor()
			visitors = append(visitors, v)
			visitorAnalyzers[name] = v
		} else {
			legacyAnalyzers = append(legacyAnalyzers, name)
		}
	}

	// Run visitors if any
	if len(visitors) > 0 {
		traverser := NewMultiAnalyzerTraverser()
		for _, v := range visitors {
			traverser.RegisterVisitor(v)
		}
		traverser.Traverse(root)

		for name, v := range visitorAnalyzers {
			combinedReport[name] = v.GetReport()
		}
	}

	// Run legacy analyzers
	for _, name := range legacyAnalyzers {
		report, err := f.RunAnalyzer(name, root)

		if err != nil {
			return nil, err
		}

		combinedReport[name] = report
	}

	return combinedReport, nil
}

func NewFactory(analyzers []StaticAnalyzer) *Factory {
	f := &Factory{
		analyzers: make(map[string]StaticAnalyzer),
	}

	for _, a := range analyzers {
		f.RegisterAnalyzer(a)
	}

	return f
}
