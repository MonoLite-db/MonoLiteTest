// Created by Yanjunhui

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

// 命令行参数 // EN: Command line arguments
var (
	mode       = flag.String("mode", "api", "测试模式: api 或 wire")                             // EN: Test mode: api or wire
	monoDBPath = flag.String("monodb", "../../testdata/fixtures/test.monodb", "MonoLite 数据库文件") // EN: MonoLite database file
	testCases  = flag.String("testcases", "../../testdata/fixtures/testcases.json", "测试用例文件")   // EN: Test cases file
	output     = flag.String("output", "../../reports/go_results.json", "结果输出文件")              // EN: Result output file
	wirePort   = flag.Int("port", 27018, "Wire Protocol 服务端口")                                 // EN: Wire Protocol server port
)

// main 主函数
// EN: main is the entry point of the program.
func main() {
	flag.Parse()

	log.Printf("=== Go 测试运行器 (%s 模式) ===", *mode) // EN: Go test runner (%s mode)

	// 加载测试用例 // EN: Load test cases
	suite, err := loadTestCases(*testCases)
	if err != nil {
		log.Fatalf("加载测试用例失败: %v", err) // EN: Failed to load test cases
	}
	log.Printf("加载了 %d 个测试用例", len(suite.Tests)) // EN: Loaded %d test cases

	var results []TestResult
	var passed, failed int

	switch *mode {
	case "api":
		results, passed, failed = runAPITests(suite)
	case "wire":
		results, passed, failed = runWireTests(suite)
	default:
		log.Fatalf("未知模式: %s", *mode) // EN: Unknown mode
	}

	// 保存结果 // EN: Save results
	resultsFile := ResultsFile{
		Language: "go",
		Mode:     *mode,
		Results:  results,
		Summary: Summary{
			Total:  len(results),
			Passed: passed,
			Failed: failed,
		},
	}

	if err := saveResults(*output, resultsFile); err != nil {
		log.Fatalf("保存结果失败: %v", err) // EN: Failed to save results
	}

	log.Printf("=== 测试完成 ===")                                     // EN: Test completed
	log.Printf("通过: %d, 失败: %d, 总计: %d", passed, failed, len(results)) // EN: Passed: %d, Failed: %d, Total: %d
	log.Printf("结果已保存到: %s", *output)                                // EN: Results saved to
}

// loadTestCases 加载测试用例
// EN: loadTestCases loads test cases from a JSON file.
func loadTestCases(path string) (*TestSuite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var suite TestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, err
	}
	return &suite, nil
}

// runAPITests 运行 API 模式测试
// EN: runAPITests runs tests in API mode.
func runAPITests(suite *TestSuite) ([]TestResult, int, int) {
	runner, err := NewAPIRunner(*monoDBPath)
	if err != nil {
		log.Fatalf("创建 API 运行器失败: %v", err) // EN: Failed to create API runner
	}
	defer runner.Close()

	var results []TestResult
	passed, failed := 0, 0

	for i, tc := range suite.Tests {
		log.Printf("[%d/%d] 测试: %s", i+1, len(suite.Tests), tc.Name) // EN: [%d/%d] Test: %s
		result := runner.RunTest(tc)
		results = append(results, result)

		if result.Success {
			passed++
			log.Printf("  ✓ 通过 (%dms)", result.Duration) // EN: Passed
		} else {
			failed++
			log.Printf("  ✗ 失败: %s (%dms)", result.Error, result.Duration) // EN: Failed
		}
	}

	return results, passed, failed
}

// runWireTests 运行 Wire 模式测试
// EN: runWireTests runs tests in Wire protocol mode.
func runWireTests(suite *TestSuite) ([]TestResult, int, int) {
	runner, err := NewWireRunner(*monoDBPath, *wirePort)
	if err != nil {
		log.Fatalf("创建 Wire 运行器失败: %v", err) // EN: Failed to create Wire runner
	}
	defer runner.Close()

	var results []TestResult
	passed, failed := 0, 0

	for i, tc := range suite.Tests {
		log.Printf("[%d/%d] 测试: %s", i+1, len(suite.Tests), tc.Name) // EN: [%d/%d] Test: %s
		result := runner.RunTest(tc)
		results = append(results, result)

		if result.Success {
			passed++
			log.Printf("  ✓ 通过 (%dms)", result.Duration) // EN: Passed
		} else {
			failed++
			log.Printf("  ✗ 失败: %s (%dms)", result.Error, result.Duration) // EN: Failed
		}
	}

	return results, passed, failed
}

// saveResults 保存测试结果
// EN: saveResults saves test results to a JSON file.
func saveResults(path string, results ResultsFile) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化结果失败: %w", err) // EN: Failed to serialize results
	}
	return os.WriteFile(path, data, 0644)
}
