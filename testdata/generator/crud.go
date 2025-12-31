// Created by Yanjunhui

package main

// GenerateCRUDTests 生成 CRUD 测试用例
// EN: GenerateCRUDTests generates CRUD test cases.
func GenerateCRUDTests() []TestCase {
	var tests []TestCase

	// Insert 测试 // EN: Insert tests
	tests = append(tests, generateInsertTests()...)

	// Find 测试 // EN: Find tests
	tests = append(tests, generateFindTests()...)

	// Update 测试 // EN: Update tests
	tests = append(tests, generateUpdateTests()...)

	// Delete 测试 // EN: Delete tests
	tests = append(tests, generateDeleteTests()...)

	// ReplaceOne 测试 // EN: ReplaceOne tests
	tests = append(tests, generateReplaceTests()...)

	// FindAndModify 测试 // EN: FindAndModify tests
	tests = append(tests, generateFindAndModifyTests()...)

	// Distinct 测试 // EN: Distinct tests
	tests = append(tests, generateDistinctTests()...)

	return tests
}

// generateInsertTests 生成插入测试用例
// EN: generateInsertTests generates insert test cases.
func generateInsertTests() []TestCase {
	return []TestCase{
		{
			Name:        "insert_single_doc",
			Category:    "crud",
			Operation:   "insert",
			Collection:  "crud_test",
			Description: "插入单个文档", // EN: Insert a single document
			Action: TestAction{
				Method: "insertOne",
				Doc:    doc("name", "Alice", "age", 25),
			},
			Expected: Expected{Count: intPtr(1)},
		},
		{
			Name:        "insert_multiple_docs",
			Category:    "crud",
			Operation:   "insert",
			Collection:  "crud_test",
			Description: "批量插入多个文档", // EN: Insert multiple documents in batch
			Action: TestAction{
				Method: "insertMany",
				Docs: []any{
					doc("name", "Bob", "age", 30),
					doc("name", "Carol", "age", 28),
					doc("name", "David", "age", 35),
				},
			},
			Expected: Expected{Count: intPtr(3)},
		},
		{
			Name:        "insert_with_custom_id",
			Category:    "crud",
			Operation:   "insert",
			Collection:  "crud_test",
			Description: "使用自定义 _id 插入", // EN: Insert with custom _id
			Action: TestAction{
				Method: "insertOne",
				Doc:    doc("_id", "custom_id_001", "value", "test"),
			},
			Expected: Expected{Count: intPtr(1)},
		},
		{
			Name:        "insert_nested_doc",
			Category:    "crud",
			Operation:   "insert",
			Collection:  "crud_test",
			Description: "插入嵌套文档", // EN: Insert nested document
			Action: TestAction{
				Method: "insertOne",
				Doc: doc(
					"_id", "nested_001",
					"user", doc("name", "Eve", "profile", doc("age", 22, "city", "Shanghai")),
				),
			},
			Expected: Expected{Count: intPtr(1)},
		},
		{
			Name:        "insert_with_array",
			Category:    "crud",
			Operation:   "insert",
			Collection:  "crud_test",
			Description: "插入包含数组的文档", // EN: Insert document with array
			Action: TestAction{
				Method: "insertOne",
				Doc:    doc("_id", "array_001", "tags", []any{"go", "swift", "typescript"}),
			},
			Expected: Expected{Count: intPtr(1)},
		},
	}
}

// generateFindTests 生成查询测试用例
// EN: generateFindTests generates find test cases.
func generateFindTests() []TestCase {
	return []TestCase{
		{
			Name:        "find_all",
			Category:    "crud",
			Operation:   "find",
			Collection:  "base",
			Description: "查询所有文档", // EN: Find all documents
			Action: TestAction{
				Method: "find",
				Filter: doc(),
			},
			Expected: Expected{Count: intPtr(8)}, // base collection has 8 docs // EN: base collection has 8 docs
		},
		{
			Name:        "find_by_id",
			Category:    "crud",
			Operation:   "find",
			Collection:  "base",
			Description: "按 _id 查询", // EN: Find by _id
			Action: TestAction{
				Method: "findOne",
				Filter: doc("_id", "base_001"),
			},
			Expected: Expected{Count: intPtr(1)},
		},
		{
			Name:        "find_with_filter",
			Category:    "crud",
			Operation:   "find",
			Collection:  "base",
			Description: "带条件查询", // EN: Find with filter
			Action: TestAction{
				Method: "find",
				Filter: doc("type", "string"),
			},
			Expected: Expected{Count: intPtr(1)},
		},
		{
			Name:        "find_with_sort",
			Category:    "crud",
			Operation:   "find",
			Collection:  "base",
			Description: "带排序查询", // EN: Find with sort
			Action: TestAction{
				Method: "find",
				Filter: doc(),
				Options: doc(
					"sort", doc("_id", -1),
					"limit", 3,
				),
			},
			Expected: Expected{Count: intPtr(3)},
		},
		{
			Name:        "find_with_projection",
			Category:    "crud",
			Operation:   "find",
			Collection:  "base",
			Description: "带投影查询", // EN: Find with projection
			Action: TestAction{
				Method: "find",
				Filter: doc(),
				Options: doc(
					"projection", doc("type", 1, "_id", 0),
					"limit", 2,
				),
			},
			Expected: Expected{Count: intPtr(2)},
		},
		{
			Name:        "find_with_skip",
			Category:    "crud",
			Operation:   "find",
			Collection:  "base",
			Description: "带跳过的分页查询", // EN: Find with skip for pagination
			Action: TestAction{
				Method: "find",
				Filter: doc(),
				Options: doc(
					"skip", 2,
					"limit", 3,
				),
			},
			Expected: Expected{Count: intPtr(3)},
		},
	}
}

// generateUpdateTests 生成更新测试用例
// EN: generateUpdateTests generates update test cases.
func generateUpdateTests() []TestCase {
	return []TestCase{
		{
			Name:        "update_single_doc",
			Category:    "crud",
			Operation:   "update",
			Collection:  "crud_test",
			Description: "更新单个文档", // EN: Update a single document
			Setup: []SetupStep{
				{Operation: "insert", Data: doc("_id", "update_001", "name", "Frank", "age", 40)},
			},
			Action: TestAction{
				Method: "updateOne",
				Filter: doc("_id", "update_001"),
				Update: doc("$set", doc("age", 41)),
			},
			Expected: Expected{MatchedCount: intPtr(1), ModifiedCount: intPtr(1)},
		},
		{
			Name:        "update_with_upsert",
			Category:    "crud",
			Operation:   "update",
			Collection:  "crud_test",
			Description: "Upsert 操作", // EN: Upsert operation
			Action: TestAction{
				Method: "updateOne",
				Filter: doc("_id", "upsert_001"),
				Update: doc("$set", doc("name", "George", "created", true)),
				Options: doc("upsert", true),
			},
			Expected: Expected{MatchedCount: intPtr(0), ModifiedCount: intPtr(0)},
		},
		{
			Name:        "update_multiple_docs",
			Category:    "crud",
			Operation:   "update",
			Collection:  "crud_test",
			Description: "更新多个文档", // EN: Update multiple documents
			Setup: []SetupStep{
				{Operation: "insert", Data: doc("_id", "multi_001", "status", "active")},
				{Operation: "insert", Data: doc("_id", "multi_002", "status", "active")},
				{Operation: "insert", Data: doc("_id", "multi_003", "status", "active")},
			},
			Action: TestAction{
				Method: "updateMany",
				Filter: doc("status", "active"),
				Update: doc("$set", doc("status", "inactive")),
			},
			Expected: Expected{MatchedCount: intPtr(3), ModifiedCount: intPtr(3)},
		},
	}
}

// generateDeleteTests 生成删除测试用例
// EN: generateDeleteTests generates delete test cases.
func generateDeleteTests() []TestCase {
	return []TestCase{
		{
			Name:        "delete_single_doc",
			Category:    "crud",
			Operation:   "delete",
			Collection:  "crud_test",
			Description: "删除单个文档", // EN: Delete a single document
			Setup: []SetupStep{
				{Operation: "insert", Data: doc("_id", "delete_001", "temp", true)},
			},
			Action: TestAction{
				Method: "deleteOne",
				Filter: doc("_id", "delete_001"),
			},
			Expected: Expected{DeletedCount: intPtr(1)},
		},
		{
			Name:        "delete_multiple_docs",
			Category:    "crud",
			Operation:   "delete",
			Collection:  "crud_test",
			Description: "删除多个文档", // EN: Delete multiple documents
			Setup: []SetupStep{
				{Operation: "insert", Data: doc("_id", "del_multi_001", "toDelete", true)},
				{Operation: "insert", Data: doc("_id", "del_multi_002", "toDelete", true)},
			},
			Action: TestAction{
				Method: "deleteMany",
				Filter: doc("toDelete", true),
			},
			Expected: Expected{DeletedCount: intPtr(2)},
		},
		{
			Name:        "delete_no_match",
			Category:    "crud",
			Operation:   "delete",
			Collection:  "crud_test",
			Description: "删除不存在的文档", // EN: Delete non-existent document
			Action: TestAction{
				Method: "deleteOne",
				Filter: doc("_id", "non_existent_id"),
			},
			Expected: Expected{DeletedCount: intPtr(0)},
		},
	}
}

// generateReplaceTests 生成替换测试用例
// EN: generateReplaceTests generates replace test cases.
func generateReplaceTests() []TestCase {
	return []TestCase{
		{
			Name:        "replace_single_doc",
			Category:    "crud",
			Operation:   "replace",
			Collection:  "crud_test",
			Description: "替换单个文档", // EN: Replace a single document
			Setup: []SetupStep{
				{Operation: "insert", Data: doc("_id", "replace_001", "old", "value")},
			},
			Action: TestAction{
				Method: "replaceOne",
				Filter: doc("_id", "replace_001"),
				Doc:    doc("_id", "replace_001", "new", "value", "replaced", true),
			},
			Expected: Expected{MatchedCount: intPtr(1), ModifiedCount: intPtr(1)},
		},
	}
}

// generateFindAndModifyTests 生成 findAndModify 测试用例
// EN: generateFindAndModifyTests generates findAndModify test cases.
func generateFindAndModifyTests() []TestCase {
	return []TestCase{
		{
			Name:        "find_and_modify_update",
			Category:    "crud",
			Operation:   "findAndModify",
			Collection:  "crud_test",
			Description: "findAndModify 更新", // EN: findAndModify update
			Setup: []SetupStep{
				{Operation: "insert", Data: doc("_id", "fam_001", "counter", 0)},
			},
			Action: TestAction{
				Method: "findAndModify",
				Filter: doc("_id", "fam_001"),
				Update: doc("$inc", doc("counter", 1)),
				Options: doc("new", true),
			},
			Expected: Expected{Count: intPtr(1)},
		},
		{
			Name:        "find_and_modify_delete",
			Category:    "crud",
			Operation:   "findAndModify",
			Collection:  "crud_test",
			Description: "findAndModify 删除", // EN: findAndModify delete
			Setup: []SetupStep{
				{Operation: "insert", Data: doc("_id", "fam_002", "toRemove", true)},
			},
			Action: TestAction{
				Method: "findAndModify",
				Filter: doc("_id", "fam_002"),
				Options: doc("remove", true),
			},
			Expected: Expected{Count: intPtr(1)},
		},
	}
}

// generateDistinctTests 生成 distinct 测试用例
// EN: generateDistinctTests generates distinct test cases.
func generateDistinctTests() []TestCase {
	return []TestCase{
		{
			Name:        "distinct_simple",
			Category:    "crud",
			Operation:   "distinct",
			Collection:  "base",
			Description: "简单 distinct 查询", // EN: Simple distinct query
			Action: TestAction{
				Method: "distinct",
				Options: doc("field", "type"),
			},
			Expected: Expected{Count: intPtr(8)}, // 8 unique types // EN: 8 unique types
		},
	}
}
