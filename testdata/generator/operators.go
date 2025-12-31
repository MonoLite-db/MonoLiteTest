// Created by Yanjunhui

package main

// GenerateUpdateOperatorTests 生成更新操作符测试
// EN: GenerateUpdateOperatorTests generates update operator test cases.
func GenerateUpdateOperatorTests() []TestCase {
	return []TestCase{
		{
			Name:        "update_op_set",
			Category:    "update_op",
			Operation:   "update",
			Collection:  "op_test",
			Description: "$set 操作符", // EN: $set operator
			Setup: []SetupStep{
				{Operation: "insert", Data: doc("_id", "set_001", "name", "test")},
			},
			Action: TestAction{
				Method: "updateOne",
				Filter: doc("_id", "set_001"),
				Update: doc("$set", doc("name", "updated")),
			},
			Expected: Expected{MatchedCount: intPtr(1), ModifiedCount: intPtr(1)},
		},
		{
			Name:        "update_op_inc",
			Category:    "update_op",
			Operation:   "update",
			Collection:  "op_test",
			Description: "$inc 操作符", // EN: $inc operator
			Setup: []SetupStep{
				{Operation: "insert", Data: doc("_id", "inc_001", "count", 10)},
			},
			Action: TestAction{
				Method: "updateOne",
				Filter: doc("_id", "inc_001"),
				Update: doc("$inc", doc("count", 5)),
			},
			Expected: Expected{MatchedCount: intPtr(1), ModifiedCount: intPtr(1)},
		},
		{
			Name:        "update_op_push",
			Category:    "update_op",
			Operation:   "update",
			Collection:  "op_test",
			Description: "$push 操作符", // EN: $push operator
			Setup: []SetupStep{
				{Operation: "insert", Data: doc("_id", "push_001", "items", []any{"a", "b"})},
			},
			Action: TestAction{
				Method: "updateOne",
				Filter: doc("_id", "push_001"),
				Update: doc("$push", doc("items", "c")),
			},
			Expected: Expected{MatchedCount: intPtr(1), ModifiedCount: intPtr(1)},
		},
	}
}

// GenerateQueryOperatorTests 生成查询操作符测试
// EN: GenerateQueryOperatorTests generates query operator test cases.
func GenerateQueryOperatorTests() []TestCase {
	return []TestCase{
		{
			Name:        "query_op_eq",
			Category:    "query_op",
			Operation:   "find",
			Collection:  "base",
			Description: "$eq 操作符", // EN: $eq operator
			Action: TestAction{
				Method: "find",
				Filter: doc("type", doc("$eq", "string")),
			},
			Expected: Expected{Count: intPtr(1)},
		},
		{
			Name:        "query_op_gt",
			Category:    "query_op",
			Operation:   "find",
			Collection:  "op_test",
			Description: "$gt 操作符", // EN: $gt operator
			Setup: []SetupStep{
				{Operation: "insert", Data: doc("_id", "num_001", "value", 10)},
				{Operation: "insert", Data: doc("_id", "num_002", "value", 20)},
				{Operation: "insert", Data: doc("_id", "num_003", "value", 30)},
			},
			Action: TestAction{
				Method: "find",
				Filter: doc("value", doc("$gt", 15)),
			},
			Expected: Expected{Count: intPtr(2)},
		},
		{
			Name:        "query_op_in",
			Category:    "query_op",
			Operation:   "find",
			Collection:  "base",
			Description: "$in 操作符", // EN: $in operator
			Action: TestAction{
				Method: "find",
				Filter: doc("type", doc("$in", []any{"string", "int32"})),
			},
			Expected: Expected{Count: intPtr(2)},
		},
	}
}
