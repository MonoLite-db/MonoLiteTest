// Created by Yanjunhui

import { MongoClient, Db, Collection } from 'mongodb';
import { Database, ProtocolServer } from 'monolite';
import { TestCase, TestResult } from './types';
import { deserialize } from 'bson';

export class WireRunner {
  private db: Database | null = null;
  private server: ProtocolServer | null = null;
  private client: MongoClient | null = null;
  private mongoDb: Db | null = null;
  private dbPath: string;
  private port: number;

  constructor(dbPath: string, port: number = 27019) {
    this.dbPath = dbPath;
    this.port = port;
  }

  async open(): Promise<void> {
    // 打开 MonoLite 数据库
    this.db = await Database.open({ filePath: this.dbPath });

    // 启动 Wire Protocol 服务器
    this.server = new ProtocolServer(`localhost:${this.port}`, this.db);
    await this.server.start();

    // 等待服务器启动
    await new Promise((resolve) => setTimeout(resolve, 100));

    // 连接 MongoDB 客户端
    this.client = new MongoClient(`mongodb://localhost:${this.port}`, {
      directConnection: true,
    });
    await this.client.connect();
    this.mongoDb = this.client.db('test');
  }

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

      // 执行前置步骤
      await this.executeSetup(col, tc);

      // 执行测试动作
      await this.executeAction(col, tc, result);

      result.success = true;
    } catch (err: any) {
      const errorMsg = err.message || String(err);
      result.error = errorMsg;
      if (tc.expected.error && errorMsg.includes(tc.expected.error)) {
        result.success = true;
      }
    }

    result.duration_ms = Date.now() - start;
    return result;
  }

  private async executeSetup(col: Collection, tc: TestCase): Promise<void> {
    if (!tc.setup || tc.setup.length === 0) return;

    for (const step of tc.setup) {
      const data = this.parseData(step.data);
      switch (step.operation) {
        case 'insert':
          await col.insertOne(data);
          break;
        case 'createIndex':
          const keys = data.keys || {};
          const opts = data.options || {};
          await col.createIndex(keys, opts);
          break;
      }
    }
  }

  private async executeAction(
    col: Collection,
    tc: TestCase,
    result: TestResult
  ): Promise<void> {
    const action = tc.action;
    const filter = action.filter ? this.parseData(action.filter) : {};
    const update = action.update ? this.parseData(action.update) : {};
    const doc = action.doc ? this.parseData(action.doc) : {};
    const docs = action.docs ? this.parseData(action.docs) : [];
    const options = action.options ? this.parseData(action.options) : {};

    switch (action.method) {
      case 'insertOne': {
        await col.insertOne(doc);
        result.count = 1;
        break;
      }

      case 'insertMany': {
        const res = await col.insertMany(docs);
        result.count = res.insertedCount;
        break;
      }

      case 'find': {
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
        const upsert = options.upsert || false;
        const res = await col.updateOne(filter, update, { upsert });
        result.matched_count = res.matchedCount;
        result.modified_count = res.modifiedCount;
        result.upserted_id = res.upsertedId;
        break;
      }

      case 'updateMany': {
        const res = await col.updateMany(filter, update);
        result.matched_count = res.matchedCount;
        result.modified_count = res.modifiedCount;
        break;
      }

      case 'deleteOne': {
        const res = await col.deleteOne(filter);
        result.deleted_count = res.deletedCount;
        break;
      }

      case 'deleteMany': {
        const res = await col.deleteMany(filter);
        result.deleted_count = res.deletedCount;
        break;
      }

      case 'replaceOne': {
        const res = await col.replaceOne(filter, doc);
        result.matched_count = res.matchedCount;
        result.modified_count = res.modifiedCount;
        break;
      }

      case 'distinct': {
        const field = options.field || '';
        const values = await col.distinct(field, filter);
        result.count = values.length;
        break;
      }

      case 'aggregate': {
        const pipeline = options.pipeline || [];
        const cursor = col.aggregate(pipeline);
        const docs = await cursor.toArray();
        result.count = docs.length;
        result.documents = docs;
        break;
      }

      case 'createIndex': {
        const keys = options.keys || {};
        const indexOpts = options.options || {};
        const name = await col.createIndex(keys, indexOpts);
        result.count = 1;
        result.documents = [{ indexName: name }];
        break;
      }

      case 'listIndexes': {
        const indexes = await col.listIndexes().toArray();
        result.count = indexes.length;
        break;
      }

      case 'dropIndex': {
        const name = options.name || '';
        await col.dropIndex(name);
        break;
      }

      default:
        throw new Error(`Unknown method: ${action.method}`);
    }
  }

  private parseData(data: any): any {
    if (data === null || data === undefined) return data;

    if (Buffer.isBuffer(data) || data instanceof Uint8Array) {
      return deserialize(Buffer.from(data));
    }

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
