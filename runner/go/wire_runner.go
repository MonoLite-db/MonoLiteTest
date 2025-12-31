// Created by Yanjunhui

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/monolite/monodb/engine"
	"github.com/monolite/monodb/protocol"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// WireRunner 通过 Wire Protocol 测试
// EN: WireRunner tests through Wire Protocol.
type WireRunner struct {
	db     *engine.Database   // 数据库实例 // EN: Database instance
	server *protocol.Server   // Wire Protocol 服务器 // EN: Wire Protocol server
	client *mongo.Client      // MongoDB 客户端 // EN: MongoDB client
	addr   string             // 服务器地址 // EN: Server address
}

// NewWireRunner 创建 Wire 运行器
// EN: NewWireRunner creates a Wire runner.
func NewWireRunner(dbPath string, port int) (*WireRunner, error) {
	db, err := engine.OpenDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err) // EN: Failed to open database
	}

	addr := fmt.Sprintf(":%d", port)
	server := protocol.NewServer(addr, db)

	// 启动服务器 // EN: Start server
	go func() {
		if err := server.Start(); err != nil {
			fmt.Printf("服务器启动失败: %v\n", err) // EN: Server start failed
		}
	}()

	// 等待服务器启动 // EN: Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// 连接客户端 // EN: Connect client
	ctx := context.Background()
	clientOpts := options.Client().
		ApplyURI(fmt.Sprintf("mongodb://localhost:%d", port)).
		SetDirect(true)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("连接客户端失败: %w", err) // EN: Failed to connect client
	}

	return &WireRunner{
		db:     db,
		server: server,
		client: client,
		addr:   addr,
	}, nil
}

// Close 关闭连接
// EN: Close closes all connections.
func (r *WireRunner) Close() error {
	ctx := context.Background()
	if r.client != nil {
		r.client.Disconnect(ctx)
	}
	if r.server != nil {
		r.server.Stop()
	}
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// RunTest 运行单个测试
// EN: RunTest runs a single test case.
func (r *WireRunner) RunTest(tc TestCase) TestResult {
	start := time.Now()
	result := TestResult{
		TestName: tc.Name,
		Language: "go",
		Mode:     "wire",
	}

	ctx := context.Background()
	col := r.client.Database("test").Collection(tc.Collection)

	// 执行前置步骤 // EN: Execute setup steps
	if err := r.executeSetup(ctx, col, tc); err != nil {
		result.Error = fmt.Sprintf("Setup 失败: %v", err) // EN: Setup failed
		result.Duration = time.Since(start).Milliseconds()
		return result
	}

	// 执行测试动作 // EN: Execute test action
	if err := r.executeAction(ctx, col, tc, &result); err != nil {
		result.Error = err.Error()
		result.Duration = time.Since(start).Milliseconds()
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
func (r *WireRunner) executeSetup(ctx context.Context, col *mongo.Collection, tc TestCase) error {
	if len(tc.Setup) == 0 {
		return nil
	}

	for _, step := range tc.Setup {
		switch step.Operation {
		case "insert":
			doc := toBsonD(step.Data)
			if _, err := col.InsertOne(ctx, doc); err != nil {
				return err
			}
		case "createIndex":
			opts := toBsonD(step.Data)
			keysRaw := getField(opts, "keys")
			keys := toBsonD(keysRaw)
			indexModel := mongo.IndexModel{Keys: keys}
			if indexOpts := getFieldD(opts, "options"); indexOpts != nil {
				if v := getField(indexOpts, "unique"); v != nil {
					unique := v.(bool)
					indexModel.Options = options.Index().SetUnique(unique)
				}
				if v := getField(indexOpts, "name"); v != nil {
					if indexModel.Options == nil {
						indexModel.Options = options.Index()
					}
					indexModel.Options.SetName(v.(string))
				}
			}
			if _, err := col.Indexes().CreateOne(ctx, indexModel); err != nil {
				return err
			}
		}
	}
	return nil
}

// executeAction 执行测试动作
// EN: executeAction executes the test action.
func (r *WireRunner) executeAction(ctx context.Context, col *mongo.Collection, tc TestCase, result *TestResult) error {
	switch tc.Action.Method {
	case "insertOne":
		return r.executeInsertOne(ctx, col, tc, result)
	case "insertMany":
		return r.executeInsertMany(ctx, col, tc, result)
	case "find":
		return r.executeFind(ctx, col, tc, result)
	case "findOne":
		return r.executeFindOne(ctx, col, tc, result)
	case "updateOne":
		return r.executeUpdateOne(ctx, col, tc, result)
	case "updateMany":
		return r.executeUpdateMany(ctx, col, tc, result)
	case "deleteOne":
		return r.executeDeleteOne(ctx, col, tc, result)
	case "deleteMany":
		return r.executeDeleteMany(ctx, col, tc, result)
	case "replaceOne":
		return r.executeReplaceOne(ctx, col, tc, result)
	case "aggregate":
		return r.executeAggregate(ctx, col, tc, result)
	case "distinct":
		return r.executeDistinct(ctx, col, tc, result)
	case "createIndex":
		return r.executeCreateIndex(ctx, col, tc, result)
	case "listIndexes":
		return r.executeListIndexes(ctx, col, tc, result)
	case "dropIndex":
		return r.executeDropIndex(ctx, col, tc, result)
	default:
		return fmt.Errorf("未知方法: %s", tc.Action.Method) // EN: Unknown method
	}
}

// executeInsertOne 执行插入单个文档
// EN: executeInsertOne executes insert one document.
func (r *WireRunner) executeInsertOne(ctx context.Context, col *mongo.Collection, tc TestCase, result *TestResult) error {
	doc := toBsonD(tc.Action.Doc)
	_, err := col.InsertOne(ctx, doc)
	if err != nil {
		return err
	}
	result.Count = 1
	return nil
}

// executeInsertMany 执行批量插入文档
// EN: executeInsertMany executes insert many documents.
func (r *WireRunner) executeInsertMany(ctx context.Context, col *mongo.Collection, tc TestCase, result *TestResult) error {
	docs := toBsonDSlice(tc.Action.Docs)
	ifaces := make([]interface{}, len(docs))
	for i, d := range docs {
		ifaces[i] = d
	}
	res, err := col.InsertMany(ctx, ifaces)
	if err != nil {
		return err
	}
	result.Count = int64(len(res.InsertedIDs))
	return nil
}

// executeFind 执行查询
// EN: executeFind executes a find query.
func (r *WireRunner) executeFind(ctx context.Context, col *mongo.Collection, tc TestCase, result *TestResult) error {
	var filter bson.D
	if tc.Action.Filter != nil {
		filter = toBsonD(tc.Action.Filter)
	}

	findOpts := options.Find()
	if tc.Action.Options != nil {
		opts := toBsonD(tc.Action.Options)
		if v := getField(opts, "sort"); v != nil {
			findOpts.SetSort(toBsonD(v))
		}
		if v := getField(opts, "limit"); v != nil {
			findOpts.SetLimit(toInt64(v))
		}
		if v := getField(opts, "skip"); v != nil {
			findOpts.SetSkip(toInt64(v))
		}
		if v := getField(opts, "projection"); v != nil {
			findOpts.SetProjection(toBsonD(v))
		}
	}

	cursor, err := col.Find(ctx, filter, findOpts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var docs []bson.M
	if err := cursor.All(ctx, &docs); err != nil {
		return err
	}
	result.Count = int64(len(docs))
	result.Documents = docs
	return nil
}

// executeFindOne 执行查询单个文档
// EN: executeFindOne executes a find one query.
func (r *WireRunner) executeFindOne(ctx context.Context, col *mongo.Collection, tc TestCase, result *TestResult) error {
	filter := toBsonD(tc.Action.Filter)

	var doc bson.M
	err := col.FindOne(ctx, filter).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		result.Count = 0
		return nil
	}
	if err != nil {
		return err
	}
	result.Count = 1
	result.Documents = []bson.M{doc}
	return nil
}

// executeUpdateOne 执行更新单个文档
// EN: executeUpdateOne executes update one document.
func (r *WireRunner) executeUpdateOne(ctx context.Context, col *mongo.Collection, tc TestCase, result *TestResult) error {
	filter := toBsonD(tc.Action.Filter)
	update := toBsonD(tc.Action.Update)

	updateOpts := options.Update()
	if tc.Action.Options != nil {
		opts := toBsonD(tc.Action.Options)
		if v := getField(opts, "upsert"); v != nil {
			updateOpts.SetUpsert(v.(bool))
		}
	}

	res, err := col.UpdateOne(ctx, filter, update, updateOpts)
	if err != nil {
		return err
	}
	result.MatchedCount = res.MatchedCount
	result.ModifiedCount = res.ModifiedCount
	result.UpsertedID = res.UpsertedID
	return nil
}

// executeUpdateMany 执行更新多个文档
// EN: executeUpdateMany executes update many documents.
func (r *WireRunner) executeUpdateMany(ctx context.Context, col *mongo.Collection, tc TestCase, result *TestResult) error {
	filter := toBsonD(tc.Action.Filter)
	update := toBsonD(tc.Action.Update)

	res, err := col.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}
	result.MatchedCount = res.MatchedCount
	result.ModifiedCount = res.ModifiedCount
	return nil
}

// executeDeleteOne 执行删除单个文档
// EN: executeDeleteOne executes delete one document.
func (r *WireRunner) executeDeleteOne(ctx context.Context, col *mongo.Collection, tc TestCase, result *TestResult) error {
	filter := toBsonD(tc.Action.Filter)
	res, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	result.DeletedCount = res.DeletedCount
	return nil
}

// executeDeleteMany 执行删除多个文档
// EN: executeDeleteMany executes delete many documents.
func (r *WireRunner) executeDeleteMany(ctx context.Context, col *mongo.Collection, tc TestCase, result *TestResult) error {
	filter := toBsonD(tc.Action.Filter)
	res, err := col.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	result.DeletedCount = res.DeletedCount
	return nil
}

// executeReplaceOne 执行替换单个文档
// EN: executeReplaceOne executes replace one document.
func (r *WireRunner) executeReplaceOne(ctx context.Context, col *mongo.Collection, tc TestCase, result *TestResult) error {
	filter := toBsonD(tc.Action.Filter)
	replacement := toBsonD(tc.Action.Doc)
	res, err := col.ReplaceOne(ctx, filter, replacement)
	if err != nil {
		return err
	}
	result.MatchedCount = res.MatchedCount
	result.ModifiedCount = res.ModifiedCount
	return nil
}

// executeAggregate 执行聚合管道
// EN: executeAggregate executes aggregation pipeline.
func (r *WireRunner) executeAggregate(ctx context.Context, col *mongo.Collection, tc TestCase, result *TestResult) error {
	opts := toBsonD(tc.Action.Options)
	pipelineRaw := getField(opts, "pipeline")

	// 转换为 bson.A // EN: Convert to bson.A
	var pipeline bson.A
	switch p := pipelineRaw.(type) {
	case []interface{}:
		for _, stage := range p {
			pipeline = append(pipeline, toBsonD(stage))
		}
	case bson.A:
		for _, stage := range p {
			pipeline = append(pipeline, toBsonD(stage))
		}
	default:
		return fmt.Errorf("pipeline 类型错误: %T", pipelineRaw) // EN: Pipeline type error
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var docs []bson.M
	if err := cursor.All(ctx, &docs); err != nil {
		return err
	}
	result.Count = int64(len(docs))
	result.Documents = docs
	return nil
}

// executeDistinct 执行去重查询
// EN: executeDistinct executes distinct query.
func (r *WireRunner) executeDistinct(ctx context.Context, col *mongo.Collection, tc TestCase, result *TestResult) error {
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

	values, err := col.Distinct(ctx, field, filter)
	if err != nil {
		return err
	}
	result.Count = int64(len(values))
	return nil
}

// executeCreateIndex 执行创建索引
// EN: executeCreateIndex executes create index operation.
func (r *WireRunner) executeCreateIndex(ctx context.Context, col *mongo.Collection, tc TestCase, result *TestResult) error {
	opts := toBsonD(tc.Action.Options)

	keysRaw := getField(opts, "keys")
	keys := toBsonD(keysRaw)
	indexModel := mongo.IndexModel{Keys: keys}

	if indexOpts := getFieldD(opts, "options"); indexOpts != nil {
		indexModel.Options = options.Index()
		if v := getField(indexOpts, "unique"); v != nil {
			indexModel.Options.SetUnique(v.(bool))
		}
		if v := getField(indexOpts, "name"); v != nil {
			indexModel.Options.SetName(v.(string))
		}
	}

	name, err := col.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}
	result.Count = 1
	result.Documents = []bson.M{{"indexName": name}}
	return nil
}

// executeListIndexes 执行列出索引
// EN: executeListIndexes executes list indexes operation.
func (r *WireRunner) executeListIndexes(ctx context.Context, col *mongo.Collection, tc TestCase, result *TestResult) error {
	cursor, err := col.Indexes().List(ctx)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		return err
	}
	result.Count = int64(len(indexes))
	return nil
}

// executeDropIndex 执行删除索引
// EN: executeDropIndex executes drop index operation.
func (r *WireRunner) executeDropIndex(ctx context.Context, col *mongo.Collection, tc TestCase, result *TestResult) error {
	opts := toBsonD(tc.Action.Options)
	name := getField(opts, "name").(string)
	_, err := col.Indexes().DropOne(ctx, name)
	return err
}
