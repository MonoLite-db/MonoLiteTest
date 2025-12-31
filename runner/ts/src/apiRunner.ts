// Created by Yanjunhui

import { Database } from 'monolite';
import { TestCase, TestResult } from './types';
import { deserialize } from 'bson';

/**
 * API 测试运行器类
 * 直接调用 MonoLite API 执行测试
 * // EN: API test runner class
 * // EN: Executes tests by directly calling MonoLite API
 */
export class APIRunner {
  /** MonoLite 数据库实例 // EN: MonoLite database instance */
  private db: Database | null = null;
  /** 数据库文件路径 // EN: Database file path */
  private dbPath: string;

  /**
   * 构造函数
   * // EN: Constructor
   * @param dbPath 数据库文件路径 // EN: Database file path
   */
  constructor(dbPath: string) {
    this.dbPath = dbPath;
  }

  /**
   * 打开数据库连接
   * // EN: Open database connection
   */
  async open(): Promise<void> {
    this.db = await Database.open({ filePath: this.dbPath });
  }

  /**
   * 关闭数据库连接
   * // EN: Close database connection
   */
  async close(): Promise<void> {
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
      mode: 'api',
      success: false,
      duration_ms: 0,
    };

    try {
      // 执行前置步骤 // EN: Execute setup steps
      await this.executeSetup(tc);

      // 执行测试动作 // EN: Execute test action
      await this.executeAction(tc, result);

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
   * @param tc 测试用例 // EN: Test case
   */
  private async executeSetup(tc: TestCase): Promise<void> {
    if (!tc.setup || tc.setup.length === 0) return;
    if (!this.db) throw new Error('Database not opened');

    // 获取集合（如果不存在则创建） // EN: Get collection (create if not exists)
    const col = await this.db.getCollection(tc.collection, true);
    if (!col) throw new Error(`Failed to get collection: ${tc.collection}`);

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
   * @param tc 测试用例 // EN: Test case
   * @param result 测试结果对象 // EN: Test result object
   */
  private async executeAction(tc: TestCase, result: TestResult): Promise<void> {
    if (!this.db) throw new Error('Database not opened');

    // 获取集合（如果不存在则创建） // EN: Get collection (create if not exists)
    const col = await this.db.getCollection(tc.collection, true);
    if (!col) throw new Error(`Failed to get collection: ${tc.collection}`);

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
        const res = await col.insertOne(doc);
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
        const findOpts: any = {};
        if (options.sort) findOpts.sort = options.sort;
        if (options.limit) findOpts.limit = options.limit;
        if (options.skip) findOpts.skip = options.skip;
        if (options.projection) findOpts.projection = options.projection;

        const results = await col.find({ filter, ...findOpts });
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
        const res = await col.updateOne(filter, update, upsert);
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

      case 'findAndModify': {
        // 查找并修改文档 // EN: Find and modify document
        const cmd: any = {
          findAndModify: tc.collection,
          query: filter,
        };
        if (update && Object.keys(update).length > 0) {
          cmd.update = update;
        }
        if (options.new !== undefined) {
          cmd.new = options.new;
        }
        if (options.upsert !== undefined) {
          cmd.upsert = options.upsert;
        }
        if (options.remove !== undefined) {
          cmd.remove = options.remove;
        }
        const res = await this.db!.runCommand(cmd);
        if (res.value) {
          result.count = 1;
          result.documents = [res.value];
        } else {
          result.count = 0;
        }
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
        const cmd = { aggregate: tc.collection, pipeline, cursor: {} };
        const res = await this.db!.runCommand(cmd);
        const cursor = res.cursor as any;
        const docs = cursor?.firstBatch || [];
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
        const indexes = col.listIndexes();
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
