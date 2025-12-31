// Created by Yanjunhui

package main

import (
	"fmt"
	"time"

	"github.com/monolite/monodb/engine"
	"go.mongodb.org/mongo-driver/bson"
)

// APIRunner 使用库 API 直接测试
// EN: APIRunner tests using the library API directly.
type APIRunner struct {
	db *engine.Database // 数据库实例 // EN: Database instance
}

// NewAPIRunner 创建 API 运行器
// EN: NewAPIRunner creates an API runner.
func NewAPIRunner(dbPath string) (*APIRunner, error) {
	db, err := engine.OpenDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err) // EN: Failed to open database
	}
	return &APIRunner{db: db}, nil
}

// Close 关闭数据库
// EN: Close closes the database connection.
func (r *APIRunner) Close() error {
	return r.db.Close()
}

// RunTest 运行单个测试
// EN: RunTest runs a single test case.
func (r *APIRunner) RunTest(tc TestCase) TestResult {
	start := time.Now()
	result := TestResult{
		TestName: tc.Name,
		Language: "go",
		Mode:     "api",
	}

	// 执行前置步骤 // EN: Execute setup steps
	if err := r.executeSetup(tc); err != nil {
		result.Error = fmt.Sprintf("Setup 失败: %v", err) // EN: Setup failed
		result.Duration = time.Since(start).Milliseconds()
		return result
	}

	// 执行测试动作 // EN: Execute test action
	if err := r.executeAction(tc, &result); err != nil {
		result.Error = err.Error()
		result.Duration = time.Since(start).Milliseconds()
		// 检查是否是预期的错误 // EN: Check if this is an expected error
		if tc.Expected.Error != "" {
			result.Success = true
		}
		return result
	}

	result.Success = true
	result.Duration = time.Since(start).Milliseconds()
	return result
}

// executeSetup 执行前置步骤
// EN: executeSetup executes the setup steps.
func (r *APIRunner) executeSetup(tc TestCase) error {
	if len(tc.Setup) == 0 {
		return nil
	}

	col, err := r.db.Collection(tc.Collection)
	if err != nil {
		return err
	}

	for _, step := range tc.Setup {
		switch step.Operation {
		case "insert":
			doc := toBsonD(step.Data)
			if _, err := col.Insert(doc); err != nil {
				return fmt.Errorf("插入失败: %w", err) // EN: Insert failed
			}
		case "createIndex":
			opts := toBsonD(step.Data)
			keys := getField(opts, "keys").(bson.D)
			indexOpts := getFieldD(opts, "options")
			if _, err := col.CreateIndex(keys, indexOpts); err != nil {
				return fmt.Errorf("创建索引失败: %w", err) // EN: Create index failed
			}
		}
	}
	return nil
}

// executeAction 执行测试动作
// EN: executeAction executes the test action.
func (r *APIRunner) executeAction(tc TestCase, result *TestResult) error {
	col, err := r.db.Collection(tc.Collection)
	if err != nil {
		return err
	}

	switch tc.Action.Method {
	case "insertOne":
		return r.executeInsertOne(col, tc, result)
	case "insertMany":
		return r.executeInsertMany(col, tc, result)
	case "find":
		return r.executeFind(col, tc, result)
	case "findOne":
		return r.executeFindOne(col, tc, result)
	case "updateOne":
		return r.executeUpdateOne(col, tc, result)
	case "updateMany":
		return r.executeUpdateMany(col, tc, result)
	case "deleteOne":
		return r.executeDeleteOne(col, tc, result)
	case "deleteMany":
		return r.executeDeleteMany(col, tc, result)
	case "replaceOne":
		return r.executeReplaceOne(col, tc, result)
	case "findAndModify":
		return r.executeFindAndModify(col, tc, result)
	case "distinct":
		return r.executeDistinct(col, tc, result)
	case "aggregate":
		return r.executeAggregate(col, tc, result)
	case "createIndex":
		return r.executeCreateIndex(col, tc, result)
	case "listIndexes":
		return r.executeListIndexes(col, tc, result)
	case "dropIndex":
		return r.executeDropIndex(col, tc, result)
	default:
		return fmt.Errorf("未知方法: %s", tc.Action.Method) // EN: Unknown method
	}
}

// executeInsertOne 执行插入单个文档
// EN: executeInsertOne executes insert one document.
func (r *APIRunner) executeInsertOne(col *engine.Collection, tc TestCase, result *TestResult) error {
	doc := toBsonD(tc.Action.Doc)
	ids, err := col.Insert(doc)
	if err != nil {
		return err
	}
	result.Count = int64(len(ids))
	return nil
}

// executeInsertMany 执行批量插入文档
// EN: executeInsertMany executes insert many documents.
func (r *APIRunner) executeInsertMany(col *engine.Collection, tc TestCase, result *TestResult) error {
	docs := toBsonDSlice(tc.Action.Docs)
	ids, err := col.Insert(docs...)
	if err != nil {
		return err
	}
	result.Count = int64(len(ids))
	return nil
}

// executeFind 执行查询
// EN: executeFind executes a find query.
func (r *APIRunner) executeFind(col *engine.Collection, tc TestCase, result *TestResult) error {
	var filter bson.D
	if tc.Action.Filter != nil {
		filter = toBsonD(tc.Action.Filter)
	}

	var opts *engine.QueryOptions
	if tc.Action.Options != nil {
		opts = &engine.QueryOptions{}
		optDoc := toBsonD(tc.Action.Options)
		if v := getField(optDoc, "sort"); v != nil {
			opts.Sort = toBsonD(v)
		}
		if v := getField(optDoc, "limit"); v != nil {
			opts.Limit = toInt64(v)
		}
		if v := getField(optDoc, "skip"); v != nil {
			opts.Skip = toInt64(v)
		}
		if v := getField(optDoc, "projection"); v != nil {
			opts.Projection = toBsonD(v)
		}
	}

	var docs []bson.D
	var err error
	if opts != nil {
		docs, err = col.FindWithOptions(filter, opts)
	} else {
		docs, err = col.Find(filter)
	}
	if err != nil {
		return err
	}

	result.Count = int64(len(docs))
	result.Documents = toMaps(docs)
	return nil
}

// executeFindOne 执行查询单个文档
// EN: executeFindOne executes a find one query.
func (r *APIRunner) executeFindOne(col *engine.Collection, tc TestCase, result *TestResult) error {
	filter := toBsonD(tc.Action.Filter)
	doc, err := col.FindOne(filter)
	if err != nil {
		return err
	}
	if doc != nil {
		result.Count = 1
		result.Documents = []bson.M{toMap(doc)}
	} else {
		result.Count = 0
	}
	return nil
}

// executeUpdateOne 执行更新单个文档
// EN: executeUpdateOne executes update one document.
func (r *APIRunner) executeUpdateOne(col *engine.Collection, tc TestCase, result *TestResult) error {
	filter := toBsonD(tc.Action.Filter)
	update := toBsonD(tc.Action.Update)

	upsert := false
	if tc.Action.Options != nil {
		opts := toBsonD(tc.Action.Options)
		if v := getField(opts, "upsert"); v != nil {
			upsert = v.(bool)
		}
	}

	// 使用 Update 方法但限制为单个文档 // EN: Use Update method but limit to single document
	updateResult, err := col.Update(filter, update, upsert)
	if err != nil {
		return err
	}
	result.MatchedCount = updateResult.MatchedCount
	result.ModifiedCount = updateResult.ModifiedCount
	result.UpsertedID = updateResult.UpsertedID
	return nil
}

// executeUpdateMany 执行更新多个文档
// EN: executeUpdateMany executes update many documents.
func (r *APIRunner) executeUpdateMany(col *engine.Collection, tc TestCase, result *TestResult) error {
	filter := toBsonD(tc.Action.Filter)
	update := toBsonD(tc.Action.Update)

	updateResult, err := col.Update(filter, update, false)
	if err != nil {
		return err
	}
	result.MatchedCount = updateResult.MatchedCount
	result.ModifiedCount = updateResult.ModifiedCount
	return nil
}

// executeDeleteOne 执行删除单个文档
// EN: executeDeleteOne executes delete one document.
func (r *APIRunner) executeDeleteOne(col *engine.Collection, tc TestCase, result *TestResult) error {
	filter := toBsonD(tc.Action.Filter)
	count, err := col.DeleteOne(filter)
	if err != nil {
		return err
	}
	result.DeletedCount = count
	return nil
}

// executeDeleteMany 执行删除多个文档
// EN: executeDeleteMany executes delete many documents.
func (r *APIRunner) executeDeleteMany(col *engine.Collection, tc TestCase, result *TestResult) error {
	filter := toBsonD(tc.Action.Filter)
	count, err := col.Delete(filter)
	if err != nil {
		return err
	}
	result.DeletedCount = count
	return nil
}

// executeReplaceOne 执行替换单个文档
// EN: executeReplaceOne executes replace one document.
func (r *APIRunner) executeReplaceOne(col *engine.Collection, tc TestCase, result *TestResult) error {
	filter := toBsonD(tc.Action.Filter)
	replacement := toBsonD(tc.Action.Doc)
	count, err := col.ReplaceOne(filter, replacement)
	if err != nil {
		return err
	}
	result.MatchedCount = count
	result.ModifiedCount = count
	return nil
}

// executeFindAndModify 执行查找并修改
// EN: executeFindAndModify executes find and modify operation.
func (r *APIRunner) executeFindAndModify(col *engine.Collection, tc TestCase, result *TestResult) error {
	opts := &engine.FindAndModifyOptions{}

	if tc.Action.Filter != nil {
		opts.Query = toBsonD(tc.Action.Filter)
	}

	if tc.Action.Update != nil {
		opts.Update = toBsonD(tc.Action.Update)
	}

	if tc.Action.Options != nil {
		optDoc := toBsonD(tc.Action.Options)
		if v := getField(optDoc, "new"); v != nil {
			opts.New = v.(bool)
		}
		if v := getField(optDoc, "upsert"); v != nil {
			opts.Upsert = v.(bool)
		}
		if v := getField(optDoc, "remove"); v != nil {
			opts.Remove = v.(bool)
		}
	}

	doc, err := col.FindAndModify(opts)
	if err != nil {
		return err
	}
	if doc != nil {
		result.Count = 1
		result.Documents = []bson.M{toMap(doc)}
	}
	return nil
}

// executeDistinct 执行去重查询
// EN: executeDistinct executes distinct query.
func (r *APIRunner) executeDistinct(col *engine.Collection, tc TestCase, result *TestResult) error {
	var filter bson.D
	if tc.Action.Filter != nil {
		filter = toBsonD(tc.Action.Filter)
	}

	var field string
	if tc.Action.Options != nil {
		opts := toBsonD(tc.Action.Options)
		if v := getField(opts, "field"); v != nil {
			field = v.(string)
		}
	}

	values, err := col.Distinct(field, filter)
	if err != nil {
		return err
	}
	result.Count = int64(len(values))
	return nil
}

// executeAggregate 执行聚合管道
// EN: executeAggregate executes aggregation pipeline.
func (r *APIRunner) executeAggregate(col *engine.Collection, tc TestCase, result *TestResult) error {
	opts := toBsonD(tc.Action.Options)

	pipelineRaw := getField(opts, "pipeline")
	if pipelineRaw == nil {
		return fmt.Errorf("缺少 pipeline") // EN: Missing pipeline
	}

	// 转换为 []bson.D // EN: Convert to []bson.D
	var stages []bson.D
	switch p := pipelineRaw.(type) {
	case []interface{}:
		for _, stage := range p {
			stages = append(stages, toBsonD(stage))
		}
	case bson.A:
		for _, stage := range p {
			stages = append(stages, toBsonD(stage))
		}
	default:
		return fmt.Errorf("pipeline 类型错误: %T", pipelineRaw) // EN: Pipeline type error
	}

	docs, err := col.Aggregate(stages)
	if err != nil {
		return err
	}
	result.Count = int64(len(docs))
	result.Documents = toMaps(docs)
	return nil
}

// executeCreateIndex 执行创建索引
// EN: executeCreateIndex executes create index operation.
func (r *APIRunner) executeCreateIndex(col *engine.Collection, tc TestCase, result *TestResult) error {
	opts := toBsonD(tc.Action.Options)

	keysRaw := getField(opts, "keys")
	keys := toBsonD(keysRaw)
	indexOpts := getFieldD(opts, "options")

	name, err := col.CreateIndex(keys, indexOpts)
	if err != nil {
		return err
	}
	result.Count = 1
	result.Documents = []bson.M{{"indexName": name}}
	return nil
}

// executeListIndexes 执行列出索引
// EN: executeListIndexes executes list indexes operation.
func (r *APIRunner) executeListIndexes(col *engine.Collection, tc TestCase, result *TestResult) error {
	indexes := col.ListIndexes()
	result.Count = int64(len(indexes))
	return nil
}

// executeDropIndex 执行删除索引
// EN: executeDropIndex executes drop index operation.
func (r *APIRunner) executeDropIndex(col *engine.Collection, tc TestCase, result *TestResult) error {
	opts := toBsonD(tc.Action.Options)
	name := getField(opts, "name").(string)
	return col.DropIndex(name)
}

// toBsonD 辅助函数: 将 any 类型转换为 bson.D
// EN: toBsonD is a helper function to convert any type to bson.D.
func toBsonD(v any) bson.D {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case bson.D:
		return val
	case map[string]interface{}:
		result := bson.D{}
		for k, v := range val {
			result = append(result, bson.E{Key: k, Value: convertValue(v)})
		}
		return result
	default:
		return nil
	}
}

// toBsonDSlice 将 []any 转换为 []bson.D
// EN: toBsonDSlice converts []any to []bson.D.
func toBsonDSlice(v []any) []bson.D {
	if v == nil {
		return nil
	}
	result := make([]bson.D, len(v))
	for i, item := range v {
		result[i] = toBsonD(item)
	}
	return result
}

// convertValue 递归转换值
// EN: convertValue recursively converts values.
func convertValue(v any) any {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case map[string]interface{}:
		return toBsonD(val)
	case []interface{}:
		result := bson.A{}
		for _, item := range val {
			result = append(result, convertValue(item))
		}
		return result
	default:
		return v
	}
}

// getField 获取文档字段
// EN: getField gets a field from a document.
func getField(doc bson.D, key string) interface{} {
	for _, e := range doc {
		if e.Key == key {
			return e.Value
		}
	}
	return nil
}

// getFieldD 获取文档字段并转换为 bson.D
// EN: getFieldD gets a field from a document and converts to bson.D.
func getFieldD(doc bson.D, key string) bson.D {
	v := getField(doc, key)
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case bson.D:
		return val
	case map[string]interface{}:
		return toBsonD(val)
	default:
		return nil
	}
}

// toInt64 转换为 int64
// EN: toInt64 converts a value to int64.
func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	default:
		return 0
	}
}

// toMap 将 bson.D 转换为 bson.M
// EN: toMap converts bson.D to bson.M.
func toMap(doc bson.D) bson.M {
	m := bson.M{}
	for _, e := range doc {
		m[e.Key] = e.Value
	}
	return m
}

// toMaps 将 []bson.D 转换为 []bson.M
// EN: toMaps converts []bson.D to []bson.M.
func toMaps(docs []bson.D) []bson.M {
	result := make([]bson.M, len(docs))
	for i, doc := range docs {
		result[i] = toMap(doc)
	}
	return result
}
