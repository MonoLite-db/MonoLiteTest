// Created by Yanjunhui

package main

// GenerateIndexTests 生成索引测试
// EN: GenerateIndexTests generates index test cases.
func GenerateIndexTests() []TestCase {
	return []TestCase{
		{
			Name:        "index_create_single_field",
			Category:    "index",
			Operation:   "createIndex",
			Collection:  "index_test",
			Description: "创建单字段索引", // EN: Create single field index
			Setup: []SetupStep{
				{Operation: "insert", Data: doc("_id", "idx_001", "email", "a@test.com")},
				{Operation: "insert", Data: doc("_id", "idx_002", "email", "b@test.com")},
			},
			Action: TestAction{
				Method:  "createIndex",
				Options: doc("keys", doc("email", 1)),
			},
			Expected: Expected{Count: intPtr(1)},
		},
		{
			Name:        "index_list",
			Category:    "index",
			Operation:   "listIndexes",
			Collection:  "index_test",
			Description: "列出索引", // EN: List indexes
			Setup: []SetupStep{
				{Operation: "insert", Data: doc("_id", "idx_list_001", "field", "value")},
			},
			Action: TestAction{
				Method: "listIndexes",
			},
			Expected: Expected{Count: intPtr(1)}, // at least _id index // EN: at least _id index
		},
	}
}
