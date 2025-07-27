package cohesion

import (
	"math"
	"slices"
)

// calculateLCOM calculates the Lack of Cohesion of Methods
func (c *CohesionAnalyzer) calculateLCOM(functions []Function) float64 {
	if len(functions) <= 1 {
		return 0.0
	}

	sharedPairs := 0
	totalPairs := 0

	for i := range functions {
		for j := i + 1; j < len(functions); j++ {
			totalPairs++
			if c.haveSharedVariables(functions[i], functions[j]) {
				sharedPairs++
			}
		}
	}

	if totalPairs == 0 {
		return 0.0
	}

	return float64(totalPairs-sharedPairs) - float64(sharedPairs)
}

// haveSharedVariables checks if two functions share any variables
func (c *CohesionAnalyzer) haveSharedVariables(fn1, fn2 Function) bool {
	for _, var1 := range fn1.Variables {
		if slices.Contains(fn2.Variables, var1) {
			return true
		}
	}
	return false
}

// calculateCohesionScore calculates a normalized cohesion score (0-1)
func (c *CohesionAnalyzer) calculateCohesionScore(lcom float64, functionCount int) float64 {
	if functionCount <= 1 {
		return 1.0
	}

	maxLCOM := float64(functionCount * (functionCount - 1) / 2)
	if maxLCOM == 0 {
		return 1.0
	}

	normalizedLCOM := lcom / maxLCOM
	cohesionScore := 1.0 - normalizedLCOM

	return math.Max(0.0, math.Min(1.0, cohesionScore))
}

// calculateFunctionCohesion calculates average function-level cohesion
func (c *CohesionAnalyzer) calculateFunctionCohesion(functions []Function) float64 {
	if len(functions) == 0 {
		return 1.0
	}

	total := 0.0
	for _, fn := range functions {
		total += fn.Cohesion
	}

	return total / float64(len(functions))
}

// calculateFunctionLevelCohesion calculates cohesion for a single function using an improved
// algorithm that doesn't penalize small, focused functions like getters, setters, or simple
// utility functions (e.g., Register functions).
//
// The algorithm uses a tiered approach:
// 1. Small functions (≤5 lines) with ≤3 variables: Perfect cohesion (1.0)
//   - These are typically focused, single-purpose functions
//   - Examples: getters, setters, simple validators, registration functions
//
// 2. Small functions with more variables: Gentle linear penalty
//   - Still considers them relatively cohesive but applies small penalty
//
// 3. Larger functions: Logarithmic scaling to avoid harsh penalties
//   - Uses variable density with logarithmic penalty curve
//   - More forgiving than linear scaling for moderately complex functions
func (c *CohesionAnalyzer) calculateFunctionLevelCohesion(fn Function) float64 {
	if fn.LineCount == 0 {
		return 1.0
	}

	if c.isSmallFunction(fn) {
		return c.calculateSmallFunctionCohesion(fn)
	}

	return c.calculateLargeFunctionCohesion(fn)
}

// isSmallFunction checks if a function is small (≤5 lines)
func (c *CohesionAnalyzer) isSmallFunction(fn Function) bool {
	return fn.LineCount <= 5
}

// calculateSmallFunctionCohesion calculates cohesion for small functions
func (c *CohesionAnalyzer) calculateSmallFunctionCohesion(fn Function) float64 {
	if len(fn.Variables) <= 3 {
		return 1.0
	}
	return math.Max(0.7, 1.0-float64(len(fn.Variables)-3)*0.1)
}

// calculateLargeFunctionCohesion calculates cohesion for larger functions
func (c *CohesionAnalyzer) calculateLargeFunctionCohesion(fn Function) float64 {
	variableDensity := float64(len(fn.Variables)) / float64(fn.LineCount)

	if variableDensity <= 0.5 {
		return 1.0 - variableDensity
	}

	penalty := math.Log2(1.0+variableDensity) / 2.0
	cohesion := 1.0 - math.Min(0.8, penalty)

	return math.Max(0.2, cohesion)
}
