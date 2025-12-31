// Created by Yanjunhui

package main

// TestCase 测试用例定义
// EN: TestCase defines a test case structure.
type TestCase struct {
	Name        string      `json:"name"`        // 测试名称 // EN: Test name
	Category    string      `json:"category"`    // 分类: crud, update_op, query_op, aggregate, index, transaction // EN: Category: crud, update_op, query_op, aggregate, index, transaction
	Operation   string      `json:"operation"`   // 操作类型 // EN: Operation type
	Collection  string      `json:"collection"`  // 集合名称 // EN: Collection name
	Description string      `json:"description"` // 描述 // EN: Description
	Setup       []SetupStep `json:"setup"`       // 前置步骤 // EN: Setup steps
	Action      TestAction  `json:"action"`      // 测试动作 // EN: Test action
	Expected    Expected    `json:"expected"`    // 预期结果 // EN: Expected result
}

// SetupStep 测试前置步骤
// EN: SetupStep defines a setup step before test execution.
type SetupStep struct {
	Operation string `json:"operation"` // insert, createIndex, etc. // EN: insert, createIndex, etc.
	Data      any    `json:"data"`      // 操作数据 // EN: Operation data
}

// TestAction 测试动作
// EN: TestAction defines the action to be performed in a test.
type TestAction struct {
	Method  string `json:"method"`            // 方法名: find, insert, update, delete, aggregate, etc. // EN: Method name: find, insert, update, delete, aggregate, etc.
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
	TestName      string `json:"test_name"`
	Language      string `json:"language"`
	Mode          string `json:"mode"` // api 或 wire // EN: api or wire
	Success       bool   `json:"success"`
	Error         string `json:"error,omitempty"`
	Duration      int64  `json:"duration_ms"`
	Documents     []any  `json:"documents,omitempty"`
	Count         int64  `json:"count,omitempty"`
	MatchedCount  int64  `json:"matched_count,omitempty"`
	ModifiedCount int64  `json:"modified_count,omitempty"`
	DeletedCount  int64  `json:"deleted_count,omitempty"`
	UpsertedID    any    `json:"upserted_id,omitempty"`
}

// TestSuite 测试套件
// EN: TestSuite defines a collection of test cases.
type TestSuite struct {
	Version   string     `json:"version"`
	Generated string     `json:"generated"`
	Tests     []TestCase `json:"tests"`
}

// intPtr 辅助函数：创建 int64 指针
// EN: intPtr is a helper function to create an int64 pointer.
func intPtr(i int64) *int64 {
	return &i
}

// doc 辅助函数：将 key-value 对转换为 map
// EN: doc is a helper function to convert key-value pairs to a map.
func doc(pairs ...any) map[string]any {
	m := make(map[string]any)
	for i := 0; i+1 < len(pairs); i += 2 {
		if key, ok := pairs[i].(string); ok {
			m[key] = pairs[i+1]
		}
	}
	return m
}
