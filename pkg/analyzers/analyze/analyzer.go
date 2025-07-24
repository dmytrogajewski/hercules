package analyze

import (
	"fmt"
	"io"

	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

type Report = map[string]any

// Thresholds represents color-coded thresholds for multiple metrics
// Structure: {"metric_name": {"red": value, "yellow": value, "green": value}}
type Thresholds = map[string]map[string]any

// CodeAnalyzer interface defines the contract for all analyzers
type CodeAnalyzer interface {
	// Core analysis methods
	Name() string
	Analyze(root *node.Node) (Report, error)
	Thresholds() Thresholds

	// Aggregation methods
	CreateAggregator() ResultAggregator

	// Formatting methods
	FormatReport(report Report, writer io.Writer) error
	FormatReportJSON(report Report, writer io.Writer) error
}

// ResultAggregator defines the interface for aggregating analyzer results
type ResultAggregator interface {
	Aggregate(results map[string]Report)
	GetResult() Report
}

type Factory struct {
	analyzers map[string]CodeAnalyzer
}

func (f *Factory) RegisterAnalyzer(analyzer CodeAnalyzer) {
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

	for _, a := range analyzers {
		report, err := f.RunAnalyzer(a, root)

		if err != nil {
			return nil, err
		}

		combinedReport[a] = report
	}

	return combinedReport, nil
}

func NewFactory(analyzers []CodeAnalyzer) *Factory {
	f := &Factory{
		analyzers: make(map[string]CodeAnalyzer),
	}

	for _, a := range analyzers {
		f.RegisterAnalyzer(a)
	}

	return f
}
