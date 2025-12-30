// Created by Yanjunhui

package main

// GenerateAggregateTests 生成聚合管道测试
func GenerateAggregateTests() []TestCase {
	setup := []SetupStep{
		{Operation: "insert", Data: doc("_id", "agg_001", "name", "Alice", "age", 25, "dept", "Engineering")},
		{Operation: "insert", Data: doc("_id", "agg_002", "name", "Bob", "age", 30, "dept", "Sales")},
		{Operation: "insert", Data: doc("_id", "agg_003", "name", "Carol", "age", 28, "dept", "Engineering")},
		{Operation: "insert", Data: doc("_id", "agg_004", "name", "David", "age", 35, "dept", "Sales")},
		{Operation: "insert", Data: doc("_id", "agg_005", "name", "Eve", "age", 22, "dept", "Engineering")},
	}

	return []TestCase{
		{
			Name:        "agg_match_simple",
			Category:    "aggregate",
			Operation:   "$match",
			Collection:  "agg_test",
			Description: "$match 简单过滤",
			Setup:       setup,
			Action: TestAction{
				Method: "aggregate",
				Options: doc("pipeline", []any{
					doc("$match", doc("dept", "Engineering")),
				}),
			},
			Expected: Expected{Count: intPtr(3)},
		},
		{
			Name:        "agg_sort_limit",
			Category:    "aggregate",
			Operation:   "$sort",
			Collection:  "agg_test",
			Description: "$sort + $limit",
			Setup:       setup,
			Action: TestAction{
				Method: "aggregate",
				Options: doc("pipeline", []any{
					doc("$sort", doc("age", -1)),
					doc("$limit", 2),
				}),
			},
			Expected: Expected{Count: intPtr(2)},
		},
		{
			Name:        "agg_group_count",
			Category:    "aggregate",
			Operation:   "$group",
			Collection:  "agg_test",
			Description: "$group 分组计数",
			Setup:       setup,
			Action: TestAction{
				Method: "aggregate",
				Options: doc("pipeline", []any{
					doc("$group", doc("_id", "$dept", "count", doc("$sum", 1))),
				}),
			},
			Expected: Expected{Count: intPtr(2)}, // 2 departments
		},
		{
			Name:        "agg_project",
			Category:    "aggregate",
			Operation:   "$project",
			Collection:  "agg_test",
			Description: "$project 投影",
			Setup:       setup,
			Action: TestAction{
				Method: "aggregate",
				Options: doc("pipeline", []any{
					doc("$project", doc("name", 1, "_id", 0)),
					doc("$limit", 3),
				}),
			},
			Expected: Expected{Count: intPtr(3)},
		},
	}
}
