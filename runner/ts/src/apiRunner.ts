// Created by Yanjunhui

import { Database } from 'monolite';
import { TestCase, TestResult } from './types';
import { deserialize } from 'bson';

export class APIRunner {
  private db: Database | null = null;
  private dbPath: string;

  constructor(dbPath: string) {
    this.dbPath = dbPath;
  }

  async open(): Promise<void> {
    this.db = await Database.open({ filePath: this.dbPath });
  }

  async close(): Promise<void> {
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
      mode: 'api',
      success: false,
      duration_ms: 0,
    };

    try {
      // 执行前置步骤
      await this.executeSetup(tc);

      // 执行测试动作
      await this.executeAction(tc, result);

      result.success = true;
    } catch (err: any) {
      const errorMsg = err.message || String(err);
      result.error = errorMsg;
      // 检查是否是预期的错误
      if (tc.expected.error && errorMsg.includes(tc.expected.error)) {
        result.success = true;
      }
    }

    result.duration_ms = Date.now() - start;
    return result;
  }

  private async executeSetup(tc: TestCase): Promise<void> {
    if (!tc.setup || tc.setup.length === 0) return;
    if (!this.db) throw new Error('Database not opened');

    const col = await this.db.getCollection(tc.collection, true);
    if (!col) throw new Error(`Failed to get collection: ${tc.collection}`);

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

  private async executeAction(tc: TestCase, result: TestResult): Promise<void> {
    if (!this.db) throw new Error('Database not opened');

    const col = await this.db.getCollection(tc.collection, true);
    if (!col) throw new Error(`Failed to get collection: ${tc.collection}`);

    const action = tc.action;
    const filter = action.filter ? this.parseData(action.filter) : {};
    const update = action.update ? this.parseData(action.update) : {};
    const doc = action.doc ? this.parseData(action.doc) : {};
    const docs = action.docs ? this.parseData(action.docs) : [];
    const options = action.options ? this.parseData(action.options) : {};

    switch (action.method) {
      case 'insertOne': {
        const res = await col.insertOne(doc);
        result.count = 1;
        break;
      }

      case 'insertMany': {
        const res = await col.insertMany(docs);
        result.count = res.insertedCount;
        break;
      }

      case 'find': {
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
        const res = await col.updateOne(filter, update, upsert);
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

      case 'findAndModify': {
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
        const field = options.field || '';
        const values = await col.distinct(field, filter);
        result.count = values.length;
        break;
      }

      case 'aggregate': {
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
        const keys = options.keys || {};
        const indexOpts = options.options || {};
        const name = await col.createIndex(keys, indexOpts);
        result.count = 1;
        result.documents = [{ indexName: name }];
        break;
      }

      case 'listIndexes': {
        const indexes = col.listIndexes();
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

    // 如果是 Buffer 或 Uint8Array (BSON raw)
    if (Buffer.isBuffer(data) || data instanceof Uint8Array) {
      return deserialize(Buffer.from(data));
    }

    // 如果是对象，检查是否有 $binary 格式（JSON 序列化的 BSON）
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
