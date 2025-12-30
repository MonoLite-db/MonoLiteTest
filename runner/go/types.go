// Created by Yanjunhui

package main

import (
	"go.mongodb.org/mongo-driver/bson"
)

// TestCase 测试用例定义
type TestCase struct {
	Name        string      `json:"name"`
	Category    string      `json:"category"`
	Operation   string      `json:"operation"`
	Collection  string      `json:"collection"`
	Description string      `json:"description"`
	Setup       []SetupStep `json:"setup"`
	Action      TestAction  `json:"action"`
	Expected    Expected    `json:"expected"`
}

// SetupStep 前置步骤
type SetupStep struct {
	Operation string `json:"operation"`
	Data      any    `json:"data"`
}

// TestAction 测试动作
type TestAction struct {
	Method  string `json:"method"`
	Filter  any    `json:"filter,omitempty"`
	Update  any    `json:"update,omitempty"`
	Doc     any    `json:"doc,omitempty"`
	Docs    []any  `json:"docs,omitempty"`
	Options any    `json:"options,omitempty"`
}

// Expected 预期结果
type Expected struct {
	Count         *int64 `json:"count,omitempty"`
	Documents     []any  `json:"documents,omitempty"`
	MatchedCount  *int64 `json:"matched_count,omitempty"`
	ModifiedCount *int64 `json:"modified_count,omitempty"`
	DeletedCount  *int64 `json:"deleted_count,omitempty"`
	UpsertedID    any    `json:"upserted_id,omitempty"`
	Error         string `json:"error,omitempty"`
	IndexName     string `json:"index_name,omitempty"`
}

// TestResult 测试结果
type TestResult struct {
	TestName      string   `json:"test_name"`
	Language      string   `json:"language"`
	Mode          string   `json:"mode"`
	Success       bool     `json:"success"`
	Error         string   `json:"error,omitempty"`
	Duration      int64    `json:"duration_ms"`
	Documents     []bson.M `json:"documents,omitempty"`
	Count         int64    `json:"count,omitempty"`
	MatchedCount  int64    `json:"matched_count,omitempty"`
	ModifiedCount int64    `json:"modified_count,omitempty"`
	DeletedCount  int64    `json:"deleted_count,omitempty"`
	UpsertedID    any      `json:"upserted_id,omitempty"`
}

// TestSuite 测试套件
type TestSuite struct {
	Version   string     `json:"version"`
	Generated string     `json:"generated"`
	Tests     []TestCase `json:"tests"`
}

// ResultsFile 结果文件
type ResultsFile struct {
	Language string       `json:"language"`
	Mode     string       `json:"mode"`
	Results  []TestResult `json:"results"`
	Summary  Summary      `json:"summary"`
}

// Summary 摘要
type Summary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}
