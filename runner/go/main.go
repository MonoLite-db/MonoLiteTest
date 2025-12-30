// Created by Yanjunhui

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	mode       = flag.String("mode", "api", "测试模式: api 或 wire")
	monoDBPath = flag.String("monodb", "../../testdata/fixtures/test.monodb", "MonoLite 数据库文件")
	testCases  = flag.String("testcases", "../../testdata/fixtures/testcases.json", "测试用例文件")
	output     = flag.String("output", "../../reports/go_results.json", "结果输出文件")
	wirePort   = flag.Int("port", 27018, "Wire Protocol 服务端口")
)

func main() {
	flag.Parse()

	log.Printf("=== Go 测试运行器 (%s 模式) ===", *mode)

	// 加载测试用例
	suite, err := loadTestCases(*testCases)
	if err != nil {
		log.Fatalf("加载测试用例失败: %v", err)
	}
	log.Printf("加载了 %d 个测试用例", len(suite.Tests))

	var results []TestResult
	var passed, failed int

	switch *mode {
	case "api":
		results, passed, failed = runAPITests(suite)
	case "wire":
		results, passed, failed = runWireTests(suite)
	default:
		log.Fatalf("未知模式: %s", *mode)
	}

	// 保存结果
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
		log.Fatalf("保存结果失败: %v", err)
	}

	log.Printf("=== 测试完成 ===")
	log.Printf("通过: %d, 失败: %d, 总计: %d", passed, failed, len(results))
	log.Printf("结果已保存到: %s", *output)
}

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

func runAPITests(suite *TestSuite) ([]TestResult, int, int) {
	runner, err := NewAPIRunner(*monoDBPath)
	if err != nil {
		log.Fatalf("创建 API 运行器失败: %v", err)
	}
	defer runner.Close()

	var results []TestResult
	passed, failed := 0, 0

	for i, tc := range suite.Tests {
		log.Printf("[%d/%d] 测试: %s", i+1, len(suite.Tests), tc.Name)
		result := runner.RunTest(tc)
		results = append(results, result)

		if result.Success {
			passed++
			log.Printf("  ✓ 通过 (%dms)", result.Duration)
		} else {
			failed++
			log.Printf("  ✗ 失败: %s (%dms)", result.Error, result.Duration)
		}
	}

	return results, passed, failed
}

func runWireTests(suite *TestSuite) ([]TestResult, int, int) {
	runner, err := NewWireRunner(*monoDBPath, *wirePort)
	if err != nil {
		log.Fatalf("创建 Wire 运行器失败: %v", err)
	}
	defer runner.Close()

	var results []TestResult
	passed, failed := 0, 0

	for i, tc := range suite.Tests {
		log.Printf("[%d/%d] 测试: %s", i+1, len(suite.Tests), tc.Name)
		result := runner.RunTest(tc)
		results = append(results, result)

		if result.Success {
			passed++
			log.Printf("  ✓ 通过 (%dms)", result.Duration)
		} else {
			failed++
			log.Printf("  ✗ 失败: %s (%dms)", result.Error, result.Duration)
		}
	}

	return results, passed, failed
}

func saveResults(path string, results ResultsFile) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化结果失败: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
