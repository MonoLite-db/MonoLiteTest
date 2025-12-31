// Created by Yanjunhui

package main

import (
	"go.mongodb.org/mongo-driver/bson"
)

// TestCase 测试用例定义
// EN: TestCase defines a test case structure.
type TestCase struct {
	Name        string      `json:"name"`        // 测试名称 // EN: Test name
	Category    string      `json:"category"`    // 分类 // EN: Category
	Operation   string      `json:"operation"`   // 操作类型 // EN: Operation type
	Collection  string      `json:"collection"`  // 集合名称 // EN: Collection name
	Description string      `json:"description"` // 描述 // EN: Description
	Setup       []SetupStep `json:"setup"`       // 前置步骤 // EN: Setup steps
	Action      TestAction  `json:"action"`      // 测试动作 // EN: Test action
	Expected    Expected    `json:"expected"`    // 预期结果 // EN: Expected result
}

// SetupStep 前置步骤
// EN: SetupStep defines a setup step before test execution.
type SetupStep struct {
	Operation string `json:"operation"` // 操作类型 // EN: Operation type
	Data      any    `json:"data"`      // 操作数据 // EN: Operation data
}

// TestAction 测试动作
// EN: TestAction defines the action to be performed in a test.
type TestAction struct {
	Method  string `json:"method"`            // 方法名 // EN: Method name
	Filter  any    `json:"filter,omitempty"`  // 查询条件 // EN: Query filter
	Update  any    `json:"update,omitempty"`  // 更新内容 // EN: Update content
	Doc     any    `json:"doc,omitempty"`     // 文档 // EN: Document
	Docs    []any  `json:"docs,omitempty"`    // 文档列表 // EN: Document list
	Options any    `json:"options,omitempty"` // 选项 // EN: Options
}

// Expected 预期结果
// EN: Expected defines the expected result of a test.
type Expected struct {
	Count         *int64 `json:"count,omitempty"`          // 预期数量 // EN: Expected count
	Documents     []any  `json:"documents,omitempty"`      // 预期文档 // EN: Expected documents
	MatchedCount  *int64 `json:"matched_count,omitempty"`  // 匹配数量 // EN: Matched count
	ModifiedCount *int64 `json:"modified_count,omitempty"` // 修改数量 // EN: Modified count
	DeletedCount  *int64 `json:"deleted_count,omitempty"`  // 删除数量 // EN: Deleted count
	UpsertedID    any    `json:"upserted_id,omitempty"`    // Upsert ID // EN: Upserted ID
	Error         string `json:"error,omitempty"`          // 预期错误 // EN: Expected error
	IndexName     string `json:"index_name,omitempty"`     // 索引名称 // EN: Index name
}

// TestResult 测试结果
// EN: TestResult defines the result of a test execution.
type TestResult struct {
	TestName      string   `json:"test_name"`               // 测试名称 // EN: Test name
	Language      string   `json:"language"`                // 语言 // EN: Language
	Mode          string   `json:"mode"`                    // 模式 // EN: Mode
	Success       bool     `json:"success"`                 // 是否成功 // EN: Success status
	Error         string   `json:"error,omitempty"`         // 错误信息 // EN: Error message
	Duration      int64    `json:"duration_ms"`             // 耗时（毫秒）// EN: Duration in milliseconds
	Documents     []bson.M `json:"documents,omitempty"`     // 返回的文档 // EN: Returned documents
	Count         int64    `json:"count,omitempty"`         // 数量 // EN: Count
	MatchedCount  int64    `json:"matched_count,omitempty"` // 匹配数量 // EN: Matched count
	ModifiedCount int64    `json:"modified_count,omitempty"` // 修改数量 // EN: Modified count
	DeletedCount  int64    `json:"deleted_count,omitempty"` // 删除数量 // EN: Deleted count
	UpsertedID    any      `json:"upserted_id,omitempty"`   // Upsert ID // EN: Upserted ID
}

// TestSuite 测试套件
// EN: TestSuite defines a collection of test cases.
type TestSuite struct {
	Version   string     `json:"version"`   // 版本 // EN: Version
	Generated string     `json:"generated"` // 生成时间 // EN: Generated time
	Tests     []TestCase `json:"tests"`     // 测试用例列表 // EN: Test case list
}

// ResultsFile 结果文件
// EN: ResultsFile defines the structure of a results file.
type ResultsFile struct {
	Language string       `json:"language"` // 语言 // EN: Language
	Mode     string       `json:"mode"`     // 模式 // EN: Mode
	Results  []TestResult `json:"results"`  // 结果列表 // EN: Results list
	Summary  Summary      `json:"summary"`  // 摘要 // EN: Summary
}

// Summary 摘要
// EN: Summary defines the summary of test results.
type Summary struct {
	Total   int `json:"total"`   // 总数 // EN: Total count
	Passed  int `json:"passed"`  // 通过数 // EN: Passed count
	Failed  int `json:"failed"`  // 失败数 // EN: Failed count
	Skipped int `json:"skipped"` // 跳过数 // EN: Skipped count
}
