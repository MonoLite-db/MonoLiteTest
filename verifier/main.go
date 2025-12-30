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

var (
	resultsDir = flag.String("results-dir", "../reports", "测试结果目录")
	outputJSON = flag.String("output-json", "../reports/consistency_report.json", "JSON 报告输出")
	outputMD   = flag.String("output-md", "../reports/consistency_report.md", "Markdown 报告输出")
)

func main() {
	flag.Parse()

	log.Println("=== MonoLite 一致性验证器 ===")

	// 收集所有结果
	results := collectResults(*resultsDir)

	// 生成报告
	report := generateReport(results)

	// 保存 JSON 报告
	if err := saveJSON(*outputJSON, report); err != nil {
		log.Fatalf("保存 JSON 报告失败: %v", err)
	}
	log.Printf("JSON 报告已保存到: %s", *outputJSON)

	// 保存 Markdown 报告
	if err := saveMarkdown(*outputMD, report); err != nil {
		log.Fatalf("保存 Markdown 报告失败: %v", err)
	}
	log.Printf("Markdown 报告已保存到: %s", *outputMD)

	// 打印摘要
	printSummary(report)
}

// ResultsFile 结果文件
type ResultsFile struct {
	Language string       `json:"language"`
	Mode     string       `json:"mode"`
	Results  []TestResult `json:"results"`
	Summary  Summary      `json:"summary"`
}

// TestResult 测试结果
type TestResult struct {
	TestName      string `json:"test_name"`
	Language      string `json:"language"`
	Mode          string `json:"mode"`
	Success       bool   `json:"success"`
	Error         string `json:"error,omitempty"`
	Duration      int64  `json:"duration_ms"`
	Count         int64  `json:"count,omitempty"`
	MatchedCount  int64  `json:"matched_count,omitempty"`
	ModifiedCount int64  `json:"modified_count,omitempty"`
	DeletedCount  int64  `json:"deleted_count,omitempty"`
}

// Summary 摘要
type Summary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// Report 一致性报告
type Report struct {
	Generated    string                    `json:"generated"`
	Summary      ReportSummary             `json:"summary"`
	ByCategory   map[string]CategoryStats  `json:"by_category"`
	ByLanguage   map[string]LanguageStats  `json:"by_language"`
	ByMode       map[string]ModeStats      `json:"by_mode"`
	Comparisons  []ComparisonResult        `json:"comparisons"`
	Failures     []FailureDetail           `json:"failures"`
}

// ReportSummary 报告摘要
type ReportSummary struct {
	TotalTests      int     `json:"total_tests"`
	TotalPassed     int     `json:"total_passed"`
	TotalFailed     int     `json:"total_failed"`
	ConsistencyRate float64 `json:"consistency_rate"`
}

// CategoryStats 按类别统计
type CategoryStats struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// LanguageStats 按语言统计
type LanguageStats struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// ModeStats 按模式统计
type ModeStats struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// ComparisonResult 比较结果
type ComparisonResult struct {
	TestName    string `json:"test_name"`
	GoAPI       bool   `json:"go_api"`
	GoWire      bool   `json:"go_wire"`
	SwiftAPI    bool   `json:"swift_api"`
	SwiftWire   bool   `json:"swift_wire"`
	TSAPI       bool   `json:"ts_api"`
	TSWire      bool   `json:"ts_wire"`
	Consistent  bool   `json:"consistent"`
}

// FailureDetail 失败详情
type FailureDetail struct {
	TestName string            `json:"test_name"`
	Failures map[string]string `json:"failures"` // language_mode -> error
}

func collectResults(dir string) map[string]*ResultsFile {
	results := make(map[string]*ResultsFile)

	files := []string{
		"go_api.json", "go_wire.json",
		"swift_api.json", "swift_wire.json",
		"ts_api.json", "ts_wire.json",
	}

	for _, file := range files {
		path := filepath.Join(dir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Printf("警告: 结果文件不存在: %s", path)
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("警告: 读取文件失败 %s: %v", path, err)
			continue
		}

		var rf ResultsFile
		if err := json.Unmarshal(data, &rf); err != nil {
			log.Printf("警告: 解析文件失败 %s: %v", path, err)
			continue
		}

		key := strings.TrimSuffix(file, ".json")
		results[key] = &rf
		log.Printf("加载结果: %s (%d 个测试)", key, len(rf.Results))
	}

	return results
}

func generateReport(results map[string]*ResultsFile) *Report {
	report := &Report{
		Generated:   time.Now().Format(time.RFC3339),
		ByCategory:  make(map[string]CategoryStats),
		ByLanguage:  make(map[string]LanguageStats),
		ByMode:      make(map[string]ModeStats),
		Comparisons: []ComparisonResult{},
		Failures:    []FailureDetail{},
	}

	// 收集所有测试名称
	testNames := make(map[string]bool)
	for _, rf := range results {
		for _, r := range rf.Results {
			testNames[r.TestName] = true
		}
	}

	// 创建结果映射
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
			TestName:   testName,
			GoAPI:      getSuccess(rm, "go_api"),
			GoWire:     getSuccess(rm, "go_wire"),
			SwiftAPI:   getSuccess(rm, "swift_api"),
			SwiftWire:  getSuccess(rm, "swift_wire"),
			TSAPI:      getSuccess(rm, "ts_api"),
			TSWire:     getSuccess(rm, "ts_wire"),
		}

		// 检查一致性
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

	// 按语言统计
	for key, rf := range results {
		parts := strings.Split(key, "_")
		if len(parts) < 2 {
			continue
		}
		lang := parts[0]
		mode := parts[1]

		// 语言统计
		ls := report.ByLanguage[lang]
		ls.Total += rf.Summary.Total
		ls.Passed += rf.Summary.Passed
		ls.Failed += rf.Summary.Failed
		report.ByLanguage[lang] = ls

		// 模式统计
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

func getSuccess(rm map[string]TestResult, key string) bool {
	if r, ok := rm[key]; ok {
		return r.Success
	}
	return false
}

func saveJSON(path string, report *Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func saveMarkdown(path string, report *Report) error {
	var sb strings.Builder

	sb.WriteString("# MonoLite 三语言一致性测试报告\n\n")
	sb.WriteString(fmt.Sprintf("**生成时间**: %s\n\n", report.Generated))

	// 概览
	sb.WriteString("## 测试概览\n\n")
	sb.WriteString("| 指标 | 数值 |\n")
	sb.WriteString("|------|------|\n")
	sb.WriteString(fmt.Sprintf("| 总测试数 | %d |\n", report.Summary.TotalTests))
	sb.WriteString(fmt.Sprintf("| 通过 | %d (%.1f%%) |\n", report.Summary.TotalPassed, report.Summary.ConsistencyRate))
	sb.WriteString(fmt.Sprintf("| 失败 | %d |\n", report.Summary.TotalFailed))
	sb.WriteString("\n")

	// 按语言统计
	sb.WriteString("## 按语言统计\n\n")
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

	// 按模式统计
	sb.WriteString("## 按模式统计\n\n")
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

	// 失败详情
	if len(report.Failures) > 0 {
		sb.WriteString("## 失败详情\n\n")
		for _, f := range report.Failures {
			sb.WriteString(fmt.Sprintf("### %s\n\n", f.TestName))
			for key, err := range f.Failures {
				sb.WriteString(fmt.Sprintf("- **%s**: %s\n", key, err))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("---\n\n")
	sb.WriteString("*报告由 MonoLite 一致性验证器自动生成*\n")

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func printSummary(report *Report) {
	log.Println("=== 验证完成 ===")
	log.Printf("总测试数: %d", report.Summary.TotalTests)
	log.Printf("通过: %d (%.1f%%)", report.Summary.TotalPassed, report.Summary.ConsistencyRate)
	log.Printf("失败: %d", report.Summary.TotalFailed)
}
