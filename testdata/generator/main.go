// Created by Yanjunhui

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/monolite/monodb/engine"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 命令行参数 // EN: Command line arguments
var (
	mongoURI    = flag.String("mongo-uri", "mongodb://localhost:27017", "MongoDB 连接 URI") // EN: MongoDB connection URI
	outputDir   = flag.String("output", "../fixtures", "测试固件输出目录")                         // EN: Test fixtures output directory
	monoDBPath  = flag.String("monodb", "../fixtures/test.monodb", "MonoLite 数据库文件路径")    // EN: MonoLite database file path
	dbName      = flag.String("db", "monolite_test", "测试数据库名称")                           // EN: Test database name
	skipMongoDB = flag.Bool("skip-mongo", false, "跳过 MongoDB（仅生成 MonoLite 数据）")           // EN: Skip MongoDB (generate MonoLite data only)
)

// main 主函数
// EN: main is the entry point of the program.
func main() {
	flag.Parse()
	ctx := context.Background()

	log.Println("=== MonoLite 一致性测试数据生成器 ===") // EN: MonoLite consistency test data generator
	log.Printf("MongoDB URI: %s", *mongoURI)
	log.Printf("MonoLite 文件: %s", *monoDBPath)        // EN: MonoLite file
	log.Printf("跳过 MongoDB: %v", *skipMongoDB)        // EN: Skip MongoDB

	// 确保输出目录存在 // EN: Ensure output directory exists
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("创建输出目录失败: %v", err) // EN: Failed to create output directory
	}
	if err := os.MkdirAll(filepath.Join(*outputDir, "expected"), 0755); err != nil {
		log.Fatalf("创建 expected 目录失败: %v", err) // EN: Failed to create expected directory
	}

	var mongoClient *mongo.Client
	var mongoDB *mongo.Database

	if !*skipMongoDB {
		// 连接 MongoDB // EN: Connect to MongoDB
		log.Println("连接 MongoDB...") // EN: Connecting to MongoDB...
		var err error
		mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(*mongoURI).SetDirect(true))
		if err != nil {
			log.Fatalf("连接 MongoDB 失败: %v", err) // EN: Failed to connect to MongoDB
		}
		defer mongoClient.Disconnect(ctx)

		// 测试连接 // EN: Test connection
		if err := mongoClient.Ping(ctx, nil); err != nil {
			log.Fatalf("MongoDB ping 失败: %v", err) // EN: MongoDB ping failed
		}
		log.Println("MongoDB 连接成功") // EN: MongoDB connected successfully

		// 清理现有数据 // EN: Clean existing data
		log.Println("清理现有数据...") // EN: Cleaning existing data...
		if err := mongoClient.Database(*dbName).Drop(ctx); err != nil {
			log.Printf("警告: 清理 MongoDB 数据库失败: %v", err) // EN: Warning: Failed to clean MongoDB database
		}
		mongoDB = mongoClient.Database(*dbName)
	} else {
		log.Println("跳过 MongoDB 连接") // EN: Skipping MongoDB connection
	}

	os.Remove(*monoDBPath)

	// 打开 MonoLite // EN: Open MonoLite
	log.Println("打开 MonoLite 数据库...") // EN: Opening MonoLite database...
	monoLite, err := engine.OpenDatabase(*monoDBPath)
	if err != nil {
		log.Fatalf("打开 MonoLite 失败: %v", err) // EN: Failed to open MonoLite
	}
	defer monoLite.Close()
	log.Println("MonoLite 数据库打开成功") // EN: MonoLite database opened successfully

	// 生成所有测试用例 // EN: Generate all test cases
	log.Println("生成测试用例...") // EN: Generating test cases...
	testSuite := generateAllTests()

	// 写入基础测试数据到两个数据库 // EN: Write base test data to both databases
	log.Println("写入基础测试数据...") // EN: Writing base test data...
	writeBaseData(ctx, mongoDB, monoLite)

	// 保存测试用例定义 // EN: Save test case definitions
	testCasesPath := filepath.Join(*outputDir, "testcases.json")
	log.Printf("保存测试用例定义到: %s", testCasesPath) // EN: Saving test case definitions to
	if err := saveJSON(testCasesPath, testSuite); err != nil {
		log.Fatalf("保存测试用例失败: %v", err) // EN: Failed to save test cases
	}

	// 关闭 MonoLite 以允许其他程序读取 // EN: Close MonoLite to allow other programs to read
	if err := monoLite.Close(); err != nil {
		log.Printf("警告: 关闭 MonoLite 时出错: %v", err) // EN: Warning: Error closing MonoLite
	}

	log.Println("=== 数据生成完成 ===")                       // EN: Data generation completed
	log.Printf("测试用例总数: %d", len(testSuite.Tests))       // EN: Total test cases
	log.Printf("测试用例文件: %s", testCasesPath)              // EN: Test cases file
	log.Printf("MonoLite 文件: %s", *monoDBPath)           // EN: MonoLite file
}

// generateAllTests 生成所有测试用例
// EN: generateAllTests generates all test cases.
func generateAllTests() *TestSuite {
	var tests []TestCase

	// CRUD 测试 // EN: CRUD tests
	crudTests := GenerateCRUDTests()
	tests = append(tests, crudTests...)
	log.Printf("  CRUD 测试: %d 个", len(crudTests)) // EN: CRUD tests: %d

	// 更新操作符测试 // EN: Update operator tests
	updateOpTests := GenerateUpdateOperatorTests()
	tests = append(tests, updateOpTests...)
	log.Printf("  更新操作符测试: %d 个", len(updateOpTests)) // EN: Update operator tests: %d

	// 查询操作符测试 // EN: Query operator tests
	queryOpTests := GenerateQueryOperatorTests()
	tests = append(tests, queryOpTests...)
	log.Printf("  查询操作符测试: %d 个", len(queryOpTests)) // EN: Query operator tests: %d

	// 聚合管道测试 // EN: Aggregation pipeline tests
	aggregateTests := GenerateAggregateTests()
	tests = append(tests, aggregateTests...)
	log.Printf("  聚合管道测试: %d 个", len(aggregateTests)) // EN: Aggregation pipeline tests: %d

	// 索引测试 // EN: Index tests
	indexTests := GenerateIndexTests()
	tests = append(tests, indexTests...)
	log.Printf("  索引测试: %d 个", len(indexTests)) // EN: Index tests: %d

	// 事务测试 // EN: Transaction tests
	txnTests := GenerateTransactionTests()
	tests = append(tests, txnTests...)
	log.Printf("  事务测试: %d 个", len(txnTests)) // EN: Transaction tests: %d

	return &TestSuite{
		Version:   "1.0.0",
		Generated: time.Now().Format(time.RFC3339),
		Tests:     tests,
	}
}

// writeBaseData 写入基础测试数据
// EN: writeBaseData writes base test data to both databases.
func writeBaseData(ctx context.Context, mongoDB *mongo.Database, monoLite *engine.Database) {
	// 创建一个包含各种 BSON 类型的基础集合 // EN: Create a base collection with various BSON types
	baseData := []bson.D{
		{{Key: "_id", Value: "base_001"}, {Key: "type", Value: "string"}, {Key: "value", Value: "hello world"}},
		{{Key: "_id", Value: "base_002"}, {Key: "type", Value: "int32"}, {Key: "value", Value: int32(42)}},
		{{Key: "_id", Value: "base_003"}, {Key: "type", Value: "int64"}, {Key: "value", Value: int64(9007199254740993)}},
		{{Key: "_id", Value: "base_004"}, {Key: "type", Value: "double"}, {Key: "value", Value: 3.14159}},
		{{Key: "_id", Value: "base_005"}, {Key: "type", Value: "bool"}, {Key: "value", Value: true}},
		{{Key: "_id", Value: "base_006"}, {Key: "type", Value: "null"}, {Key: "value", Value: nil}},
		{{Key: "_id", Value: "base_007"}, {Key: "type", Value: "array"}, {Key: "value", Value: bson.A{1, 2, 3}}},
		{{Key: "_id", Value: "base_008"}, {Key: "type", Value: "document"}, {Key: "value", Value: bson.D{{Key: "nested", Value: "value"}}}},
	}

	// 写入 MongoDB（如果可用）// EN: Write to MongoDB (if available)
	if mongoDB != nil {
		mongoCol := mongoDB.Collection("base")
		for _, doc := range baseData {
			if _, err := mongoCol.InsertOne(ctx, doc); err != nil {
				log.Printf("警告: MongoDB 插入失败: %v", err) // EN: Warning: MongoDB insert failed
			}
		}
	}

	// 写入 MonoLite // EN: Write to MonoLite
	monoCol, err := monoLite.Collection("base")
	if err != nil {
		log.Printf("警告: 获取 MonoLite 集合失败: %v", err) // EN: Warning: Failed to get MonoLite collection
		return
	}
	for _, doc := range baseData {
		if _, err := monoCol.Insert(doc); err != nil {
			log.Printf("警告: MonoLite 插入失败: %v", err) // EN: Warning: MonoLite insert failed
		}
	}

	log.Printf("  基础数据: %d 条文档", len(baseData)) // EN: Base data: %d documents
}

// saveJSON 保存 JSON 文件
// EN: saveJSON saves data to a JSON file.
func saveJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 序列化失败: %w", err) // EN: JSON serialization failed
	}
	return os.WriteFile(path, data, 0644)
}
