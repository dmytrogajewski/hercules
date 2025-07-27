package common

import (
	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
)

// Aggregator provides generic aggregation capabilities for analyzers
type Aggregator struct {
	metricsProcessor   *MetricsProcessor
	dataCollector      *DataCollector
	resultBuilder      *ResultBuilder
	analyzerName       string
	messageBuilder     func(float64) string
	emptyResultBuilder func() analyze.Report
}

// NewAggregator creates a new Aggregator with configurable components
func NewAggregator(
	analyzerName string,
	numericKeys, countKeys []string,
	collectionKey, identifierKey string,
	messageBuilder func(float64) string,
) *Aggregator {
	return &Aggregator{
		metricsProcessor: NewMetricsProcessor(numericKeys, countKeys),
		dataCollector:    NewDataCollector(collectionKey, identifierKey),
		resultBuilder:    NewResultBuilder(),
		analyzerName:     analyzerName,
		messageBuilder:   messageBuilder,
	}
}

// NewAggregatorWithCustomEmptyResult creates a new Aggregator with custom empty result builder
func NewAggregatorWithCustomEmptyResult(
	analyzerName string,
	numericKeys, countKeys []string,
	collectionKey, identifierKey string,
	messageBuilder func(float64) string,
	emptyResultBuilder func() analyze.Report,
) *Aggregator {
	return &Aggregator{
		metricsProcessor:   NewMetricsProcessor(numericKeys, countKeys),
		dataCollector:      NewDataCollector(collectionKey, identifierKey),
		resultBuilder:      NewResultBuilder(),
		analyzerName:       analyzerName,
		messageBuilder:     messageBuilder,
		emptyResultBuilder: emptyResultBuilder,
	}
}

// Aggregate combines multiple analysis results
func (a *Aggregator) Aggregate(results map[string]analyze.Report) {
	for _, report := range results {
		if report == nil {
			continue
		}

		a.metricsProcessor.ProcessReport(report)
		a.dataCollector.CollectFromReport(report)
	}
}

// GetResult returns the aggregated analysis result
func (a *Aggregator) GetResult() analyze.Report {
	if a.metricsProcessor.GetReportCount() == 0 {
		if a.emptyResultBuilder != nil {
			return a.emptyResultBuilder()
		}
		return a.resultBuilder.BuildEmptyResult(a.analyzerName)
	}

	averages := a.metricsProcessor.CalculateAverages()
	counts := a.metricsProcessor.GetCounts()
	collectedData := a.dataCollector.GetSortedData()

	// Build metrics map
	metrics := make(map[string]interface{})
	for key, value := range averages {
		metrics[key] = value
	}
	for key, value := range counts {
		metrics[key] = value
	}

	// Build message
	var message string
	if a.messageBuilder != nil {
		// Use the first numeric metric for message building (can be customized)
		for _, value := range averages {
			message = a.messageBuilder(value)
			break
		}
	}
	if message == "" {
		message = "Analysis completed"
	}

	return a.resultBuilder.BuildCollectionResult(
		a.analyzerName,
		a.dataCollector.GetCollectionKey(),
		collectedData,
		metrics,
		message,
	)
}

// GetMetricsProcessor returns the metrics processor
func (a *Aggregator) GetMetricsProcessor() *MetricsProcessor {
	return a.metricsProcessor
}

// GetDataCollector returns the data collector
func (a *Aggregator) GetDataCollector() *DataCollector {
	return a.dataCollector
}

// GetResultBuilder returns the result builder
func (a *Aggregator) GetResultBuilder() *ResultBuilder {
	return a.resultBuilder
}
