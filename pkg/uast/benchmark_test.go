package uast

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

// Benchmark parsing performance with different file sizes and languages
func BenchmarkParse(b *testing.B) {
	parser, err := NewParser()
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}

	testCases := []struct {
		name     string
		filename string
		content  []byte
	}{
		{
			name:     "SmallGoFile",
			filename: "main.go",
			content:  []byte("package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}\n"),
		},
		{
			name:     "MediumGoFile",
			filename: "medium.go",
			content:  generateMediumGoFile(),
		},
		{
			name:     "LargeGoFile",
			filename: "large.go",
			content:  generateLargeGoFile(),
		},
		{
			name:     "VeryLargeGoFile",
			filename: "verylarge.go",
			content:  generateVeryLargeGoFile(),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.Logf("Parsing %s (%d bytes)", tc.filename, len(tc.content))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := parser.Parse(tc.filename, tc.content)
				if err != nil {
					b.Fatalf("Parse failed: %v", err)
				}
			}
		})
	}
}

// Benchmark DSL query performance with different query types
func BenchmarkDSLQueries(b *testing.B) {
	parser, err := NewParser()
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}

	testFiles := []struct {
		name    string
		content []byte
	}{
		{"MediumGoFile", generateMediumGoFile()},
		{"LargeGoFile", generateLargeGoFile()},
		{"VeryLargeGoFile", generateVeryLargeGoFile()},
	}

	queries := []struct {
		name  string
		query string
	}{
		{"SimpleFieldAccess", ".type"},
		{"FilterByType", "filter(.type == \"FunctionDecl\")"},
		{"FilterByRole", "filter(.roles has \"Function\")"},
		{"MapChildren", "map(.children)"},
		{"Pipeline", "map(.children) |> filter(.type == \"Identifier\")"},
		{"ComplexQuery", "map(.children) |> filter(.type == \"FunctionDecl\") |> map(.children) |> filter(.type == \"Identifier\")"},
		{"ReduceCount", "reduce(count)"},
		{"Membership", "filter(.roles has \"Declaration\")"},
		{"Comparison", "filter(.type == \"Identifier\")"},
		{"NestedField", "map(.children) |> filter(.type == \"FunctionDecl\")"},
	}

	for _, tf := range testFiles {
		node, err := parser.Parse(tf.name+".go", tf.content)
		if err != nil {
			b.Fatalf("Failed to parse test file: %v", err)
		}
		for _, q := range queries {
			b.Run(tf.name+"/"+q.name, func(b *testing.B) {
				b.Logf("Query '%s' on %s (%d bytes)", q.query, tf.name, len(tf.content))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, err := node.FindDSL(q.query)
					if err != nil {
						b.Fatalf("Query failed: %v", err)
					}
				}
			})
		}
	}
}

// Benchmark tree traversal performance
func BenchmarkTreeTraversal(b *testing.B) {
	parser, err := NewParser()
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}

	testFiles := []struct {
		name    string
		content []byte
	}{
		{"LargeGoFile", generateLargeGoFile()},
		{"VeryLargeGoFile", generateVeryLargeGoFile()},
	}

	for _, tf := range testFiles {
		root, err := parser.Parse(tf.name+".go", tf.content)
		if err != nil {
			b.Fatalf("Failed to parse test file: %v", err)
		}
		b.Run(tf.name+"/PreOrderTraversal", func(b *testing.B) {
			b.Logf("Pre-order traversal on %s (%d bytes)", tf.name, len(tf.content))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				count := 0
				for n := range root.PreOrder() {
					_ = n
					count++
				}
				if count == 0 {
					b.Fatal("No nodes traversed")
				}
			}
		})
		b.Run(tf.name+"/PostOrderTraversal", func(b *testing.B) {
			b.Logf("Post-order traversal on %s (%d bytes)", tf.name, len(tf.content))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				count := 0
				root.VisitPostOrder(func(n *node.Node) {
					_ = n
					count++
				})
				if count == 0 {
					b.Fatal("No nodes traversed")
				}
			}
		})
		b.Run(tf.name+"/FindWithPredicate", func(b *testing.B) {
			b.Logf("Find nodes with predicate on %s (%d bytes)", tf.name, len(tf.content))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				results := root.Find(func(n *node.Node) bool {
					return n.Type == "FunctionDecl"
				})
				_ = results
			}
		})
	}
}

// Benchmark change detection performance
func BenchmarkChangeDetection(b *testing.B) {
	parser, err := NewParser()
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}

	testFiles := []struct {
		name   string
		before []byte
		after  []byte
	}{
		{"MediumGoFile", generateMediumGoFile(), generateModifiedGoFile()},
		{"VeryLargeGoFile", generateVeryLargeGoFile(), generateVeryLargeGoFileModified()},
	}

	for _, tf := range testFiles {
		before, err := parser.Parse(tf.name+"_before.go", tf.before)
		if err != nil {
			b.Fatalf("Failed to parse before file: %v", err)
		}
		after, err := parser.Parse(tf.name+"_after.go", tf.after)
		if err != nil {
			b.Fatalf("Failed to parse after file: %v", err)
		}
		b.Run(tf.name+"/DetectChanges", func(b *testing.B) {
			b.Logf("Detecting changes between %s before/after (%d bytes)", tf.name, len(tf.before))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				changes := DetectChanges(before, after)
				_ = changes
			}
		})
		b.Run(tf.name+"/FilterChangesByType", func(b *testing.B) {
			changes := DetectChanges(before, after)
			b.Logf("Filtering changes by type for %s (%d bytes)", tf.name, len(tf.before))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				filtered := FilterChangesByType(changes, ChangeAdded)
				_ = filtered
			}
		})
		b.Run(tf.name+"/CountChangesByType", func(b *testing.B) {
			changes := DetectChanges(before, after)
			b.Logf("Counting changes by type for %s (%d bytes)", tf.name, len(tf.before))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				counts := CountChangesByType(changes)
				_ = counts
			}
		})
	}
}

// Benchmark memory usage for large trees
func BenchmarkMemoryUsage(b *testing.B) {
	parser, err := NewParser()
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}

	files := []struct {
		name    string
		content []byte
	}{
		{"LargeGoFile", generateLargeGoFile()},
		{"VeryLargeGoFile", generateVeryLargeGoFile()},
	}

	for _, tf := range files {
		b.Run(tf.name+"/ParseMemory", func(b *testing.B) {
			var memstatsBefore, memstatsAfter runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memstatsBefore)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				runtime.GC()
				var m1, m2 runtime.MemStats
				runtime.ReadMemStats(&m1)
				node, err := parser.Parse(tf.name+".go", tf.content)
				if err != nil {
					b.Fatalf("Parse failed: %v", err)
				}
				_ = node
				runtime.ReadMemStats(&m2)
				b.Logf("Alloc: %d, TotalAlloc: %d, Sys: %d, NumGC: %d", m2.Alloc-m1.Alloc, m2.TotalAlloc-m1.TotalAlloc, m2.Sys-m1.Sys, m2.NumGC-m1.NumGC)
			}
			b.StopTimer()
			runtime.ReadMemStats(&memstatsAfter)
			fmt.Printf("MEMORY_USAGE_JSON {\"test\":\"%s/ParseMemory\",\"alloc\":%d,\"total_alloc\":%d,\"sys\":%d,\"num_gc\":%d}\n", tf.name, memstatsAfter.Alloc-memstatsBefore.Alloc, memstatsAfter.TotalAlloc-memstatsBefore.TotalAlloc, memstatsAfter.Sys-memstatsBefore.Sys, memstatsAfter.NumGC-memstatsBefore.NumGC)
		})
		b.Run(tf.name+"/QueryMemory", func(b *testing.B) {
			node, err := parser.Parse(tf.name+".go", tf.content)
			if err != nil {
				b.Fatalf("Parse failed: %v", err)
			}
			var memstatsBefore, memstatsAfter runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memstatsBefore)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				runtime.GC()
				var m1, m2 runtime.MemStats
				runtime.ReadMemStats(&m1)
				_, err := node.FindDSL("filter(.type == \"FunctionDecl\")")
				if err != nil {
					b.Fatalf("Query failed: %v", err)
				}
				runtime.ReadMemStats(&m2)
				b.Logf("Alloc: %d, TotalAlloc: %d, Sys: %d, NumGC: %d", m2.Alloc-m1.Alloc, m2.TotalAlloc-m1.TotalAlloc, m2.Sys-m1.Sys, m2.NumGC-m1.NumGC)
			}
			b.StopTimer()
			runtime.ReadMemStats(&memstatsAfter)
			fmt.Printf("MEMORY_USAGE_JSON {\"test\":\"%s/QueryMemory\",\"alloc\":%d,\"total_alloc\":%d,\"sys\":%d,\"num_gc\":%d}\n", tf.name, memstatsAfter.Alloc-memstatsBefore.Alloc, memstatsAfter.TotalAlloc-memstatsBefore.TotalAlloc, memstatsAfter.Sys-memstatsBefore.Sys, memstatsAfter.NumGC-memstatsBefore.NumGC)
		})
	}
}

// Helper functions to generate test data
func generateMediumGoFile() []byte {
	return []byte(`package main

import (
	"fmt"
	"strings"
)

type User struct {
	Name  string
	Email string
}

func (u *User) GetDisplayName() string {
	return fmt.Sprintf("%s <%s>", u.Name, u.Email)
}

func processUsers(users []User) []string {
	var results []string
	for _, user := range users {
		if strings.Contains(user.Email, "@example.com") {
			results = append(results, user.GetDisplayName())
		}
	}
	return results
}

func main() {
	users := []User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@test.org"},
	}

	displayNames := processUsers(users)
	for _, name := range displayNames {
		fmt.Println(name)
	}
}`)
}

func generateLargeGoFile() []byte {
	// Generate a larger file with more functions and complexity
	content := `package main

import (
	"fmt"
	"strings"
	"time"
	"math/rand"
)

type Config struct {
	Host     string
	Port     int
	Timeout  time.Duration
	Features map[string]bool
}

type Processor struct {
	config *Config
	cache  map[string]interface{}
}

func NewProcessor(config *Config) *Processor {
	return &Processor{
		config: config,
		cache:  make(map[string]interface{}),
	}
}

func (p *Processor) Process(data []string) []string {
	var results []string
	for _, item := range data {
		if processed := p.processItem(item); processed != "" {
			results = append(results, processed)
		}
	}
	return results
}

func (p *Processor) processItem(item string) string {
	if cached, exists := p.cache[item]; exists {
		return cached.(string)
	}

	result := strings.ToUpper(item)
	p.cache[item] = result
	return result
}

func validateConfig(config *Config) error {
	if config.Host == "" {
		return fmt.Errorf("host is required")
	}
	if config.Port <= 0 {
		return fmt.Errorf("port must be positive")
	}
	return nil
}

func generateData(count int) []string {
	data := make([]string, count)
	for i := 0; i < count; i++ {
		data[i] = fmt.Sprintf("item_%d", i)
	}
	return data
}

func filterData(data []string, predicate func(string) bool) []string {
	var filtered []string
	for _, item := range data {
		if predicate(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func transformData(data []string, transformer func(string) string) []string {
	transformed := make([]string, len(data))
	for i, item := range data {
		transformed[i] = transformer(item)
	}
	return transformed
}

func aggregateData(data []string) map[string]int {
	aggregated := make(map[string]int)
	for _, item := range data {
		aggregated[item]++
	}
	return aggregated
}

func main() {
	config := &Config{
		Host:     "localhost",
		Port:     8080,
		Timeout:  time.Second * 30,
		Features: map[string]bool{"cache": true, "logging": true},
	}

	if err := validateConfig(config); err != nil {
		fmt.Printf("Config error: %v\n", err)
		return
	}

	processor := NewProcessor(config)
	data := generateData(100)

	// Process data through multiple stages
	filtered := filterData(data, func(item string) bool {
		return strings.Contains(item, "5")
	})

	transformed := transformData(filtered, func(item string) string {
		return strings.ToUpper(item)
	})

	processed := processor.Process(transformed)
	aggregated := aggregateData(processed)

	fmt.Printf("Processed %d items\n", len(processed))
	fmt.Printf("Aggregated into %d unique items\n", len(aggregated))
}
`
	return []byte(content)
}

func generateModifiedGoFile() []byte {
	// Similar to medium but with some modifications
	return []byte(`package main

import (
	"fmt"
	"strings"
)

type User struct {
	Name  string
	Email string
	Age   int  // Added field
}

func (u *User) GetDisplayName() string {
	return fmt.Sprintf("%s <%s> (%d)", u.Name, u.Email, u.Age)  // Modified
}

func processUsers(users []User) []string {
	var results []string
	for _, user := range users {
		if strings.Contains(user.Email, "@example.com") && user.Age >= 18 {  // Modified condition
			results = append(results, user.GetDisplayName())
		}
	}
	return results
}

func main() {
	users := []User{
		{Name: "Alice", Email: "alice@example.com", Age: 25},  // Added age
		{Name: "Bob", Email: "bob@test.org", Age: 17},         // Added age
	}

	displayNames := processUsers(users)
	for _, name := range displayNames {
		fmt.Println(name)
	}
}`)
}

func generateVeryLargeGoFile() []byte {
	// Generate a very large Go file by repeating the large file 10 times
	large := string(generateLargeGoFile())
	content := "package main\n\n"
	for i := 0; i < 10; i++ {
		content += large
	}
	return []byte(content)
}

func generateVeryLargeGoFileModified() []byte {
	// Generate a modified very large Go file for change detection
	large := string(generateLargeGoFile())
	content := "package main\n\n"
	for i := 0; i < 10; i++ {
		if i == 5 {
			content += large + "\n// modification\n"
		} else {
			content += large
		}
	}
	return []byte(content)
}
