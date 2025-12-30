// Created by Yanjunhui

package main

// TestCase 测试用例定义
type TestCase struct {
	Name        string      `json:"name"`        // 测试名称
	Category    string      `json:"category"`    // 分类: crud, update_op, query_op, aggregate, index, transaction
	Operation   string      `json:"operation"`   // 操作类型
	Collection  string      `json:"collection"`  // 集合名称
	Description string      `json:"description"` // 描述
	Setup       []SetupStep `json:"setup"`       // 前置步骤
	Action      TestAction  `json:"action"`      // 测试动作
	Expected    Expected    `json:"expected"`    // 预期结果
}

// SetupStep 测试前置步骤
type SetupStep struct {
	Operation string `json:"operation"` // insert, createIndex, etc.
	Data      any    `json:"data"`      // 操作数据
}

// TestAction 测试动作
type TestAction struct {
	Method  string `json:"method"`            // 方法名: find, insert, update, delete, aggregate, etc.
	Filter  any    `json:"filter,omitempty"`  // 查询条件
	Update  any    `json:"update,omitempty"`  // 更新内容
	Doc     any    `json:"doc,omitempty"`     // 文档
	Docs    []any  `json:"docs,omitempty"`    // 文档列表
	Options any    `json:"options,omitempty"` // 选项
}

// Expected 预期结果
type Expected struct {
	Count         *int64 `json:"count,omitempty"`          // 预期数量
	Documents     []any  `json:"documents,omitempty"`      // 预期文档
	MatchedCount  *int64 `json:"matched_count,omitempty"`  // 匹配数量
	ModifiedCount *int64 `json:"modified_count,omitempty"` // 修改数量
	DeletedCount  *int64 `json:"deleted_count,omitempty"`  // 删除数量
	UpsertedID    any    `json:"upserted_id,omitempty"`    // Upsert ID
	Error         string `json:"error,omitempty"`          // 预期错误
	IndexName     string `json:"index_name,omitempty"`     // 索引名称
}

// TestResult 测试结果
type TestResult struct {
	TestName      string `json:"test_name"`
	Language      string `json:"language"`
	Mode          string `json:"mode"` // api 或 wire
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
type TestSuite struct {
	Version   string     `json:"version"`
	Generated string     `json:"generated"`
	Tests     []TestCase `json:"tests"`
}

// 辅助函数：创建 int64 指针
func intPtr(i int64) *int64 {
	return &i
}

// doc 辅助函数：将 key-value 对转换为 map
func doc(pairs ...any) map[string]any {
	m := make(map[string]any)
	for i := 0; i+1 < len(pairs); i += 2 {
		if key, ok := pairs[i].(string); ok {
			m[key] = pairs[i+1]
		}
	}
	return m
}
