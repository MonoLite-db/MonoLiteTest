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

var (
	mongoURI    = flag.String("mongo-uri", "mongodb://localhost:27017", "MongoDB 连接 URI")
	outputDir   = flag.String("output", "../fixtures", "测试固件输出目录")
	monoDBPath  = flag.String("monodb", "../fixtures/test.monodb", "MonoLite 数据库文件路径")
	dbName      = flag.String("db", "monolite_test", "测试数据库名称")
	skipMongoDB = flag.Bool("skip-mongo", false, "跳过 MongoDB（仅生成 MonoLite 数据）")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	log.Println("=== MonoLite 一致性测试数据生成器 ===")
	log.Printf("MongoDB URI: %s", *mongoURI)
	log.Printf("MonoLite 文件: %s", *monoDBPath)
	log.Printf("跳过 MongoDB: %v", *skipMongoDB)

	// 确保输出目录存在
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("创建输出目录失败: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(*outputDir, "expected"), 0755); err != nil {
		log.Fatalf("创建 expected 目录失败: %v", err)
	}

	var mongoClient *mongo.Client
	var mongoDB *mongo.Database

	if !*skipMongoDB {
		// 连接 MongoDB
		log.Println("连接 MongoDB...")
		var err error
		mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(*mongoURI).SetDirect(true))
		if err != nil {
			log.Fatalf("连接 MongoDB 失败: %v", err)
		}
		defer mongoClient.Disconnect(ctx)

		// 测试连接
		if err := mongoClient.Ping(ctx, nil); err != nil {
			log.Fatalf("MongoDB ping 失败: %v", err)
		}
		log.Println("MongoDB 连接成功")

		// 清理现有数据
		log.Println("清理现有数据...")
		if err := mongoClient.Database(*dbName).Drop(ctx); err != nil {
			log.Printf("警告: 清理 MongoDB 数据库失败: %v", err)
		}
		mongoDB = mongoClient.Database(*dbName)
	} else {
		log.Println("跳过 MongoDB 连接")
	}

	os.Remove(*monoDBPath)

	// 打开 MonoLite
	log.Println("打开 MonoLite 数据库...")
	monoLite, err := engine.OpenDatabase(*monoDBPath)
	if err != nil {
		log.Fatalf("打开 MonoLite 失败: %v", err)
	}
	defer monoLite.Close()
	log.Println("MonoLite 数据库打开成功")

	// 生成所有测试用例
	log.Println("生成测试用例...")
	testSuite := generateAllTests()

	// 写入基础测试数据到两个数据库
	log.Println("写入基础测试数据...")
	writeBaseData(ctx, mongoDB, monoLite)

	// 保存测试用例定义
	testCasesPath := filepath.Join(*outputDir, "testcases.json")
	log.Printf("保存测试用例定义到: %s", testCasesPath)
	if err := saveJSON(testCasesPath, testSuite); err != nil {
		log.Fatalf("保存测试用例失败: %v", err)
	}

	// 关闭 MonoLite 以允许其他程序读取
	if err := monoLite.Close(); err != nil {
		log.Printf("警告: 关闭 MonoLite 时出错: %v", err)
	}

	log.Println("=== 数据生成完成 ===")
	log.Printf("测试用例总数: %d", len(testSuite.Tests))
	log.Printf("测试用例文件: %s", testCasesPath)
	log.Printf("MonoLite 文件: %s", *monoDBPath)
}

func generateAllTests() *TestSuite {
	var tests []TestCase

	// CRUD 测试
	crudTests := GenerateCRUDTests()
	tests = append(tests, crudTests...)
	log.Printf("  CRUD 测试: %d 个", len(crudTests))

	// 更新操作符测试
	updateOpTests := GenerateUpdateOperatorTests()
	tests = append(tests, updateOpTests...)
	log.Printf("  更新操作符测试: %d 个", len(updateOpTests))

	// 查询操作符测试
	queryOpTests := GenerateQueryOperatorTests()
	tests = append(tests, queryOpTests...)
	log.Printf("  查询操作符测试: %d 个", len(queryOpTests))

	// 聚合管道测试
	aggregateTests := GenerateAggregateTests()
	tests = append(tests, aggregateTests...)
	log.Printf("  聚合管道测试: %d 个", len(aggregateTests))

	// 索引测试
	indexTests := GenerateIndexTests()
	tests = append(tests, indexTests...)
	log.Printf("  索引测试: %d 个", len(indexTests))

	// 事务测试
	txnTests := GenerateTransactionTests()
	tests = append(tests, txnTests...)
	log.Printf("  事务测试: %d 个", len(txnTests))

	return &TestSuite{
		Version:   "1.0.0",
		Generated: time.Now().Format(time.RFC3339),
		Tests:     tests,
	}
}

func writeBaseData(ctx context.Context, mongoDB *mongo.Database, monoLite *engine.Database) {
	// 创建一个包含各种 BSON 类型的基础集合
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

	// 写入 MongoDB（如果可用）
	if mongoDB != nil {
		mongoCol := mongoDB.Collection("base")
		for _, doc := range baseData {
			if _, err := mongoCol.InsertOne(ctx, doc); err != nil {
				log.Printf("警告: MongoDB 插入失败: %v", err)
			}
		}
	}

	// 写入 MonoLite
	monoCol, err := monoLite.Collection("base")
	if err != nil {
		log.Printf("警告: 获取 MonoLite 集合失败: %v", err)
		return
	}
	for _, doc := range baseData {
		if _, err := monoCol.Insert(doc); err != nil {
			log.Printf("警告: MonoLite 插入失败: %v", err)
		}
	}

	log.Printf("  基础数据: %d 条文档", len(baseData))
}

func saveJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
