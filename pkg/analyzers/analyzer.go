package analyzers

import (
	"fmt"

	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

type AnalyzerReport = map[string]any

// Thresholds represents color-coded thresholds for multiple metrics
// Structure: {"metric_name": {"red": value, "yellow": value, "green": value}}
type Thresholds = map[string]map[string]any

type CodeAnalyzer interface {
	Name() string
	Analyze(root *node.Node) (AnalyzerReport, error)
	Thresholds() Thresholds
}

type Factory struct {
	analyzers map[string]CodeAnalyzer
}

func (f *Factory) RegisterAnalyzer(analyzer CodeAnalyzer) {
	f.analyzers[analyzer.Name()] = analyzer
}

func (f *Factory) RunAnalyzer(name string, root *node.Node) (AnalyzerReport, error) {
	analyzer, ok := f.analyzers[name]

	if !ok {
		return nil, fmt.Errorf("no registered analyzer with name=%s", name)
	}

	return analyzer.Analyze(root)
}

func (f *Factory) RunAnalyzers(root *node.Node, analyzers []string) (map[string]AnalyzerReport, error) {
	combinedReport := map[string]AnalyzerReport{}

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
