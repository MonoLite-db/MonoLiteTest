// Created by Yanjunhui

import { MongoClient, Db, Collection } from 'mongodb';
import { Database, ProtocolServer } from 'monolite';
import { TestCase, TestResult } from './types';
import { deserialize } from 'bson';

/**
 * Wire Protocol 测试运行器类
 * 通过 MongoDB Wire Protocol 连接 MonoLite 数据库执行测试
 * // EN: Wire Protocol test runner class
 * // EN: Executes tests by connecting to MonoLite database via MongoDB Wire Protocol
 */
export class WireRunner {
  /** MonoLite 数据库实例 // EN: MonoLite database instance */
  private db: Database | null = null;
  /** Wire Protocol 服务器实例 // EN: Wire Protocol server instance */
  private server: ProtocolServer | null = null;
  /** MongoDB 客户端实例 // EN: MongoDB client instance */
  private client: MongoClient | null = null;
  /** MongoDB 数据库实例 // EN: MongoDB database instance */
  private mongoDb: Db | null = null;
  /** 数据库文件路径 // EN: Database file path */
  private dbPath: string;
  /** Wire Protocol 端口号 // EN: Wire Protocol port number */
  private port: number;

  /**
   * 构造函数
   * // EN: Constructor
   * @param dbPath 数据库文件路径 // EN: Database file path
   * @param port Wire Protocol 端口号，默认 27019 // EN: Wire Protocol port, default 27019
   */
  constructor(dbPath: string, port: number = 27019) {
    this.dbPath = dbPath;
    this.port = port;
  }

  /**
   * 打开数据库连接并启动 Wire Protocol 服务器
   * // EN: Open database connection and start Wire Protocol server
   */
  async open(): Promise<void> {
    // 打开 MonoLite 数据库 // EN: Open MonoLite database
    this.db = await Database.open({ filePath: this.dbPath });

    // 启动 Wire Protocol 服务器 // EN: Start Wire Protocol server
    this.server = new ProtocolServer(`localhost:${this.port}`, this.db);
    await this.server.start();

    // 等待服务器启动 // EN: Wait for server to start
    await new Promise((resolve) => setTimeout(resolve, 100));

    // 连接 MongoDB 客户端 // EN: Connect MongoDB client
    this.client = new MongoClient(`mongodb://localhost:${this.port}`, {
      directConnection: true,
    });
    await this.client.connect();
    this.mongoDb = this.client.db('test');
  }

  /**
   * 关闭所有连接和服务器
   * // EN: Close all connections and server
   */
  async close(): Promise<void> {
    if (this.client) {
      await this.client.close();
      this.client = null;
    }
    if (this.server) {
      await this.server.stop();
      this.server = null;
    }
    if (this.db) {
      await this.db.close();
      this.db = null;
    }
  }

  /**
   * 运行单个测试用例
   * // EN: Run a single test case
   * @param tc 测试用例 // EN: Test case
   * @returns 测试结果 // EN: Test result
   */
  async runTest(tc: TestCase): Promise<TestResult> {
    const start = Date.now();
    const result: TestResult = {
      test_name: tc.name,
      language: 'typescript',
      mode: 'wire',
      success: false,
      duration_ms: 0,
    };

    try {
      if (!this.mongoDb) throw new Error('MongoDB not connected');

      const col = this.mongoDb.collection(tc.collection);

      // 执行前置步骤 // EN: Execute setup steps
      await this.executeSetup(col, tc);

      // 执行测试动作 // EN: Execute test action
      await this.executeAction(col, tc, result);

      result.success = true;
    } catch (err: any) {
      const errorMsg = err.message || String(err);
      result.error = errorMsg;
      // 检查是否是预期的错误 // EN: Check if it's an expected error
      if (tc.expected.error && errorMsg.includes(tc.expected.error)) {
        result.success = true;
      }
    }

    result.duration_ms = Date.now() - start;
    return result;
  }

  /**
   * 执行测试前置步骤
   * // EN: Execute test setup steps
   * @param col MongoDB 集合 // EN: MongoDB collection
   * @param tc 测试用例 // EN: Test case
   */
  private async executeSetup(col: Collection, tc: TestCase): Promise<void> {
    if (!tc.setup || tc.setup.length === 0) return;

    for (const step of tc.setup) {
      const data = this.parseData(step.data);
      switch (step.operation) {
        case 'insert':
          // 插入单个文档 // EN: Insert single document
          await col.insertOne(data);
          break;
        case 'createIndex':
          // 创建索引 // EN: Create index
          const keys = data.keys || {};
          const opts = data.options || {};
          await col.createIndex(keys, opts);
          break;
      }
    }
  }

  /**
   * 执行测试动作
   * // EN: Execute test action
   * @param col MongoDB 集合 // EN: MongoDB collection
   * @param tc 测试用例 // EN: Test case
   * @param result 测试结果对象 // EN: Test result object
   */
  private async executeAction(
    col: Collection,
    tc: TestCase,
    result: TestResult
  ): Promise<void> {
    const action = tc.action;
    // 解析各种参数 // EN: Parse various parameters
    const filter = action.filter ? this.parseData(action.filter) : {};
    const update = action.update ? this.parseData(action.update) : {};
    const doc = action.doc ? this.parseData(action.doc) : {};
    const docs = action.docs ? this.parseData(action.docs) : [];
    const options = action.options ? this.parseData(action.options) : {};

    switch (action.method) {
      case 'insertOne': {
        // 插入单个文档 // EN: Insert single document
        await col.insertOne(doc);
        result.count = 1;
        break;
      }

      case 'insertMany': {
        // 批量插入文档 // EN: Insert multiple documents
        const res = await col.insertMany(docs);
        result.count = res.insertedCount;
        break;
      }

      case 'find': {
        // 查询文档 // EN: Find documents
        let cursor = col.find(filter);
        if (options.sort) cursor = cursor.sort(options.sort);
        if (options.skip) cursor = cursor.skip(options.skip);
        if (options.limit) cursor = cursor.limit(options.limit);
        if (options.projection) cursor = cursor.project(options.projection);

        const results = await cursor.toArray();
        result.count = results.length;
        result.documents = results;
        break;
      }

      case 'findOne': {
        // 查询单个文档 // EN: Find single document
        const doc = await col.findOne(filter);
        if (doc) {
          result.count = 1;
          result.documents = [doc];
        } else {
          result.count = 0;
        }
        break;
      }

      case 'updateOne': {
        // 更新单个文档 // EN: Update single document
        const upsert = options.upsert || false;
        const res = await col.updateOne(filter, update, { upsert });
        result.matched_count = res.matchedCount;
        result.modified_count = res.modifiedCount;
        result.upserted_id = res.upsertedId;
        break;
      }

      case 'updateMany': {
        // 批量更新文档 // EN: Update multiple documents
        const res = await col.updateMany(filter, update);
        result.matched_count = res.matchedCount;
        result.modified_count = res.modifiedCount;
        break;
      }

      case 'deleteOne': {
        // 删除单个文档 // EN: Delete single document
        const res = await col.deleteOne(filter);
        result.deleted_count = res.deletedCount;
        break;
      }

      case 'deleteMany': {
        // 批量删除文档 // EN: Delete multiple documents
        const res = await col.deleteMany(filter);
        result.deleted_count = res.deletedCount;
        break;
      }

      case 'replaceOne': {
        // 替换单个文档 // EN: Replace single document
        const res = await col.replaceOne(filter, doc);
        result.matched_count = res.matchedCount;
        result.modified_count = res.modifiedCount;
        break;
      }

      case 'distinct': {
        // 获取字段的去重值 // EN: Get distinct values of a field
        const field = options.field || '';
        const values = await col.distinct(field, filter);
        result.count = values.length;
        break;
      }

      case 'aggregate': {
        // 聚合查询 // EN: Aggregate query
        const pipeline = options.pipeline || [];
        const cursor = col.aggregate(pipeline);
        const docs = await cursor.toArray();
        result.count = docs.length;
        result.documents = docs;
        break;
      }

      case 'createIndex': {
        // 创建索引 // EN: Create index
        const keys = options.keys || {};
        const indexOpts = options.options || {};
        const name = await col.createIndex(keys, indexOpts);
        result.count = 1;
        result.documents = [{ indexName: name }];
        break;
      }

      case 'listIndexes': {
        // 列出索引 // EN: List indexes
        const indexes = await col.listIndexes().toArray();
        result.count = indexes.length;
        break;
      }

      case 'dropIndex': {
        // 删除索引 // EN: Drop index
        const name = options.name || '';
        await col.dropIndex(name);
        break;
      }

      default:
        // 未知方法 // EN: Unknown method
        throw new Error(`Unknown method: ${action.method}`);
    }
  }

  /**
   * 解析数据，处理 BSON 序列化格式
   * // EN: Parse data, handle BSON serialization format
   * @param data 原始数据 // EN: Raw data
   * @returns 解析后的数据 // EN: Parsed data
   */
  private parseData(data: any): any {
    if (data === null || data === undefined) return data;

    // 如果是 Buffer 或 Uint8Array（BSON 原始格式）
    // // EN: If it's Buffer or Uint8Array (BSON raw format)
    if (Buffer.isBuffer(data) || data instanceof Uint8Array) {
      return deserialize(Buffer.from(data));
    }

    // 如果是对象，检查是否有 $binary 格式（JSON 序列化的 BSON）
    // // EN: If it's an object, check for $binary format (JSON serialized BSON)
    if (typeof data === 'object') {
      if (data.$binary && data.$binary.base64) {
        const buf = Buffer.from(data.$binary.base64, 'base64');
        return deserialize(buf);
      }
      return data;
    }

    return data;
  }
}
