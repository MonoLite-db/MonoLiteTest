// Created by Yanjunhui

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 命令行参数 // EN: Command line arguments
var (
	resultsDir = flag.String("results-dir", "../reports", "测试结果目录")                            // EN: Test results directory
	outputJSON = flag.String("output-json", "../reports/consistency_report.json", "JSON 报告输出") // EN: JSON report output
	outputMD   = flag.String("output-md", "../reports/consistency_report.md", "Markdown 报告输出") // EN: Markdown report output
)

// main 主函数
// EN: main is the entry point of the program.
func main() {
	flag.Parse()

	log.Println("=== MonoLite 一致性验证器 ===") // EN: MonoLite consistency verifier

	// 收集所有结果 // EN: Collect all results
	results := collectResults(*resultsDir)

	// 生成报告 // EN: Generate report
	report := generateReport(results)

	// 保存 JSON 报告 // EN: Save JSON report
	if err := saveJSON(*outputJSON, report); err != nil {
		log.Fatalf("保存 JSON 报告失败: %v", err) // EN: Failed to save JSON report
	}
	log.Printf("JSON 报告已保存到: %s", *outputJSON) // EN: JSON report saved to

	// 保存 Markdown 报告 // EN: Save Markdown report
	if err := saveMarkdown(*outputMD, report); err != nil {
		log.Fatalf("保存 Markdown 报告失败: %v", err) // EN: Failed to save Markdown report
	}
	log.Printf("Markdown 报告已保存到: %s", *outputMD) // EN: Markdown report saved to

	// 打印摘要 // EN: Print summary
	printSummary(report)
}

// ResultsFile 结果文件
// EN: ResultsFile defines the structure of a results file.
type ResultsFile struct {
	Language string       `json:"language"` // 语言 // EN: Language
	Mode     string       `json:"mode"`     // 模式 // EN: Mode
	Results  []TestResult `json:"results"`  // 结果列表 // EN: Results list
	Summary  Summary      `json:"summary"`  // 摘要 // EN: Summary
}

// TestResult 测试结果
// EN: TestResult defines the result of a test execution.
type TestResult struct {
	TestName      string `json:"test_name"`               // 测试名称 // EN: Test name
	Language      string `json:"language"`                // 语言 // EN: Language
	Mode          string `json:"mode"`                    // 模式 // EN: Mode
	Success       bool   `json:"success"`                 // 是否成功 // EN: Success status
	Error         string `json:"error,omitempty"`         // 错误信息 // EN: Error message
	Duration      int64  `json:"duration_ms"`             // 耗时（毫秒）// EN: Duration in milliseconds
	Count         int64  `json:"count,omitempty"`         // 数量 // EN: Count
	MatchedCount  int64  `json:"matched_count,omitempty"` // 匹配数量 // EN: Matched count
	ModifiedCount int64  `json:"modified_count,omitempty"` // 修改数量 // EN: Modified count
	DeletedCount  int64  `json:"deleted_count,omitempty"` // 删除数量 // EN: Deleted count
}

// Summary 摘要
// EN: Summary defines the summary of test results.
type Summary struct {
	Total   int `json:"total"`   // 总数 // EN: Total count
	Passed  int `json:"passed"`  // 通过数 // EN: Passed count
	Failed  int `json:"failed"`  // 失败数 // EN: Failed count
	Skipped int `json:"skipped"` // 跳过数 // EN: Skipped count
}

// Report 一致性报告
// EN: Report defines the consistency report structure.
type Report struct {
	Generated   string                   `json:"generated"`    // 生成时间 // EN: Generated time
	Summary     ReportSummary            `json:"summary"`      // 摘要 // EN: Summary
	ByCategory  map[string]CategoryStats `json:"by_category"`  // 按分类统计 // EN: Statistics by category
	ByLanguage  map[string]LanguageStats `json:"by_language"`  // 按语言统计 // EN: Statistics by language
	ByMode      map[string]ModeStats     `json:"by_mode"`      // 按模式统计 // EN: Statistics by mode
	Comparisons []ComparisonResult       `json:"comparisons"`  // 比较结果 // EN: Comparison results
	Failures    []FailureDetail          `json:"failures"`     // 失败详情 // EN: Failure details
}

// ReportSummary 报告摘要
// EN: ReportSummary defines the summary section of the report.
type ReportSummary struct {
	TotalTests      int     `json:"total_tests"`      // 总测试数 // EN: Total test count
	TotalPassed     int     `json:"total_passed"`     // 通过数 // EN: Passed count
	TotalFailed     int     `json:"total_failed"`     // 失败数 // EN: Failed count
	ConsistencyRate float64 `json:"consistency_rate"` // 一致性比率 // EN: Consistency rate
}

// CategoryStats 按类别统计
// EN: CategoryStats defines statistics by category.
type CategoryStats struct {
	Total  int `json:"total"`  // 总数 // EN: Total count
	Passed int `json:"passed"` // 通过数 // EN: Passed count
	Failed int `json:"failed"` // 失败数 // EN: Failed count
}

// LanguageStats 按语言统计
// EN: LanguageStats defines statistics by language.
type LanguageStats struct {
	Total  int `json:"total"`  // 总数 // EN: Total count
	Passed int `json:"passed"` // 通过数 // EN: Passed count
	Failed int `json:"failed"` // 失败数 // EN: Failed count
}

// ModeStats 按模式统计
// EN: ModeStats defines statistics by mode.
type ModeStats struct {
	Total  int `json:"total"`  // 总数 // EN: Total count
	Passed int `json:"passed"` // 通过数 // EN: Passed count
	Failed int `json:"failed"` // 失败数 // EN: Failed count
}

// ComparisonResult 比较结果
// EN: ComparisonResult defines the comparison result for a test case.
type ComparisonResult struct {
	TestName   string `json:"test_name"`   // 测试名称 // EN: Test name
	GoAPI      bool   `json:"go_api"`      // Go API 结果 // EN: Go API result
	GoWire     bool   `json:"go_wire"`     // Go Wire 结果 // EN: Go Wire result
	SwiftAPI   bool   `json:"swift_api"`   // Swift API 结果 // EN: Swift API result
	SwiftWire  bool   `json:"swift_wire"`  // Swift Wire 结果 // EN: Swift Wire result
	TSAPI      bool   `json:"ts_api"`      // TypeScript API 结果 // EN: TypeScript API result
	TSWire     bool   `json:"ts_wire"`     // TypeScript Wire 结果 // EN: TypeScript Wire result
	Consistent bool   `json:"consistent"`  // 是否一致 // EN: Whether consistent
}

// FailureDetail 失败详情
// EN: FailureDetail defines the detail of a failed test.
type FailureDetail struct {
	TestName string            `json:"test_name"` // 测试名称 // EN: Test name
	Failures map[string]string `json:"failures"`  // 失败信息 (language_mode -> error) // EN: Failure info (language_mode -> error)
}

// collectResults 收集所有结果
// EN: collectResults collects all test results from files.
func collectResults(dir string) map[string]*ResultsFile {
	results := make(map[string]*ResultsFile)

	// 结果文件列表 // EN: Result files list
	files := []string{
		"go_api.json", "go_wire.json",
		"swift_api.json", "swift_wire.json",
		"ts_api.json", "ts_wire.json",
	}

	for _, file := range files {
		path := filepath.Join(dir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Printf("警告: 结果文件不存在: %s", path) // EN: Warning: Result file does not exist
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("警告: 读取文件失败 %s: %v", path, err) // EN: Warning: Failed to read file
			continue
		}

		var rf ResultsFile
		if err := json.Unmarshal(data, &rf); err != nil {
			log.Printf("警告: 解析文件失败 %s: %v", path, err) // EN: Warning: Failed to parse file
			continue
		}

		key := strings.TrimSuffix(file, ".json")
		results[key] = &rf
		log.Printf("加载结果: %s (%d 个测试)", key, len(rf.Results)) // EN: Loaded results: %s (%d tests)
	}

	return results
}

// generateReport 生成报告
// EN: generateReport generates the consistency report.
func generateReport(results map[string]*ResultsFile) *Report {
	report := &Report{
		Generated:   time.Now().Format(time.RFC3339),
		ByCategory:  make(map[string]CategoryStats),
		ByLanguage:  make(map[string]LanguageStats),
		ByMode:      make(map[string]ModeStats),
		Comparisons: []ComparisonResult{},
		Failures:    []FailureDetail{},
	}

	// 收集所有测试名称 // EN: Collect all test names
	testNames := make(map[string]bool)
	for _, rf := range results {
		for _, r := range rf.Results {
			testNames[r.TestName] = true
		}
	}

	// 创建结果映射 // EN: Create result mapping
	resultMap := make(map[string]map[string]TestResult) // testName -> key -> result
	for key, rf := range results {
		for _, r := range rf.Results {
			if resultMap[r.TestName] == nil {
				resultMap[r.TestName] = make(map[string]TestResult)
			}
			resultMap[r.TestName][key] = r
		}
	}

	totalPassed := 0
	totalFailed := 0

	for testName := range testNames {
		rm := resultMap[testName]

		comp := ComparisonResult{
			TestName:  testName,
			GoAPI:     getSuccess(rm, "go_api"),
			GoWire:    getSuccess(rm, "go_wire"),
			SwiftAPI:  getSuccess(rm, "swift_api"),
			SwiftWire: getSuccess(rm, "swift_wire"),
			TSAPI:     getSuccess(rm, "ts_api"),
			TSWire:    getSuccess(rm, "ts_wire"),
		}

		// 检查一致性 // EN: Check consistency
		successCount := 0
		failureCount := 0
		failures := make(map[string]string)

		for key, r := range rm {
			if r.Success {
				successCount++
			} else {
				failureCount++
				failures[key] = r.Error
			}
		}

		comp.Consistent = failureCount == 0
		report.Comparisons = append(report.Comparisons, comp)

		if comp.Consistent {
			totalPassed++
		} else {
			totalFailed++
			report.Failures = append(report.Failures, FailureDetail{
				TestName: testName,
				Failures: failures,
			})
		}
	}

	// 按语言统计 // EN: Statistics by language
	for key, rf := range results {
		parts := strings.Split(key, "_")
		if len(parts) < 2 {
			continue
		}
		lang := parts[0]
		mode := parts[1]

		// 语言统计 // EN: Language statistics
		ls := report.ByLanguage[lang]
		ls.Total += rf.Summary.Total
		ls.Passed += rf.Summary.Passed
		ls.Failed += rf.Summary.Failed
		report.ByLanguage[lang] = ls

		// 模式统计 // EN: Mode statistics
		ms := report.ByMode[mode]
		ms.Total += rf.Summary.Total
		ms.Passed += rf.Summary.Passed
		ms.Failed += rf.Summary.Failed
		report.ByMode[mode] = ms
	}

	report.Summary = ReportSummary{
		TotalTests:  len(testNames),
		TotalPassed: totalPassed,
		TotalFailed: totalFailed,
	}
	if report.Summary.TotalTests > 0 {
		report.Summary.ConsistencyRate = float64(totalPassed) / float64(report.Summary.TotalTests) * 100
	}

	return report
}

// getSuccess 获取测试成功状态
// EN: getSuccess gets the success status of a test.
func getSuccess(rm map[string]TestResult, key string) bool {
	if r, ok := rm[key]; ok {
		return r.Success
	}
	return false
}

// saveJSON 保存 JSON 报告
// EN: saveJSON saves the report as JSON.
func saveJSON(path string, report *Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// saveMarkdown 保存 Markdown 报告
// EN: saveMarkdown saves the report as Markdown.
func saveMarkdown(path string, report *Report) error {
	var sb strings.Builder

	sb.WriteString("# MonoLite 三语言一致性测试报告\n\n")              // EN: MonoLite Three-Language Consistency Test Report
	sb.WriteString(fmt.Sprintf("**生成时间**: %s\n\n", report.Generated)) // EN: Generated time

	// 概览 // EN: Overview
	sb.WriteString("## 测试概览\n\n") // EN: Test Overview
	sb.WriteString("| 指标 | 数值 |\n")
	sb.WriteString("|------|------|\n")
	sb.WriteString(fmt.Sprintf("| 总测试数 | %d |\n", report.Summary.TotalTests))                                  // EN: Total tests
	sb.WriteString(fmt.Sprintf("| 通过 | %d (%.1f%%) |\n", report.Summary.TotalPassed, report.Summary.ConsistencyRate)) // EN: Passed
	sb.WriteString(fmt.Sprintf("| 失败 | %d |\n", report.Summary.TotalFailed))                                      // EN: Failed
	sb.WriteString("\n")

	// 按语言统计 // EN: Statistics by language
	sb.WriteString("## 按语言统计\n\n") // EN: Statistics by Language
	sb.WriteString("| 语言 | 总数 | 通过 | 失败 | 通过率 |\n")
	sb.WriteString("|------|------|------|------|--------|\n")
	for lang, stats := range report.ByLanguage {
		rate := float64(0)
		if stats.Total > 0 {
			rate = float64(stats.Passed) / float64(stats.Total) * 100
		}
		sb.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %.1f%% |\n",
			lang, stats.Total, stats.Passed, stats.Failed, rate))
	}
	sb.WriteString("\n")

	// 按模式统计 // EN: Statistics by mode
	sb.WriteString("## 按模式统计\n\n") // EN: Statistics by Mode
	sb.WriteString("| 模式 | 总数 | 通过 | 失败 | 通过率 |\n")
	sb.WriteString("|------|------|------|------|--------|\n")
	for mode, stats := range report.ByMode {
		rate := float64(0)
		if stats.Total > 0 {
			rate = float64(stats.Passed) / float64(stats.Total) * 100
		}
		sb.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %.1f%% |\n",
			mode, stats.Total, stats.Passed, stats.Failed, rate))
	}
	sb.WriteString("\n")

	// 失败详情 // EN: Failure details
	if len(report.Failures) > 0 {
		sb.WriteString("## 失败详情\n\n") // EN: Failure Details
		for _, f := range report.Failures {
			sb.WriteString(fmt.Sprintf("### %s\n\n", f.TestName))
			for key, err := range f.Failures {
				sb.WriteString(fmt.Sprintf("- **%s**: %s\n", key, err))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("---\n\n")
	sb.WriteString("*报告由 MonoLite 一致性验证器自动生成*\n") // EN: Report automatically generated by MonoLite consistency verifier

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// printSummary 打印摘要
// EN: printSummary prints the summary to console.
func printSummary(report *Report) {
	log.Println("=== 验证完成 ===")                                                                 // EN: Verification completed
	log.Printf("总测试数: %d", report.Summary.TotalTests)                                            // EN: Total tests
	log.Printf("通过: %d (%.1f%%)", report.Summary.TotalPassed, report.Summary.ConsistencyRate) // EN: Passed
	log.Printf("失败: %d", report.Summary.TotalFailed)                                            // EN: Failed
}
