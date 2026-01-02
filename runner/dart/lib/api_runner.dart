// Created by Yanjunhui

import 'package:monolite/monolite.dart';
import 'types.dart';

/// API 模式测试运行器
/// EN: APIRunner tests using the library API directly.
class APIRunner {
  final Database db; // 数据库实例

  APIRunner(this.db);

  /// 创建 API 运行器
  /// EN: Creates an API runner.
  static Future<APIRunner> create(String dbPath) async {
    final db = await Database.open(dbPath);
    return APIRunner(db);
  }

  /// 关闭数据库
  /// EN: Closes the database connection.
  Future<void> close() async {
    await db.close();
  }

  /// 运行单个测试
  /// EN: Runs a single test case.
  Future<TestResult> runTest(TestCase tc) async {
    final stopwatch = Stopwatch()..start();
    final result = TestResult(
      testName: tc.name,
      language: 'dart',
      mode: 'api',
    );

    try {
      // 执行前置步骤
      await _executeSetup(tc);

      // 执行测试动作
      await _executeAction(tc, result);

      result.success = true;
    } catch (e) {
      result.error = e.toString();
      // 检查是否是预期的错误
      if (tc.expected.error != null && tc.expected.error!.isNotEmpty) {
        result.success = true;
      }
    }

    stopwatch.stop();
    result.durationMs = stopwatch.elapsedMilliseconds;
    return result;
  }

  /// 执行前置步骤
  /// EN: Executes setup steps.
  Future<void> _executeSetup(TestCase tc) async {
    if (tc.setup.isEmpty) return;

    final col = await db.collection(tc.collection);

    for (final step in tc.setup) {
      switch (step.operation) {
        case 'insert':
          final doc = toBsonDocument(step.data);
          await col.insert([doc]);
          break;
        case 'createIndex':
          final opts = step.data as Map<String, dynamic>;
          final keys = opts['keys'] as Map<String, dynamic>;
          final indexOpts = opts['options'] as Map<String, dynamic>?;
          final unique = indexOpts?['unique'] as bool? ?? false;
          await col.createIndex(
            keys.map((k, v) => MapEntry(k, (v as num).toInt())),
            unique: unique,
          );
          break;
      }
    }
  }

  /// 执行测试动作
  /// EN: Executes test action.
  Future<void> _executeAction(TestCase tc, TestResult result) async {
    final col = await db.collection(tc.collection);

    switch (tc.action.method) {
      case 'insertOne':
        await _executeInsertOne(col, tc, result);
        break;
      case 'insertMany':
        await _executeInsertMany(col, tc, result);
        break;
      case 'find':
        await _executeFind(col, tc, result);
        break;
      case 'findOne':
        await _executeFindOne(col, tc, result);
        break;
      case 'updateOne':
        await _executeUpdateOne(col, tc, result);
        break;
      case 'updateMany':
        await _executeUpdateMany(col, tc, result);
        break;
      case 'deleteOne':
        await _executeDeleteOne(col, tc, result);
        break;
      case 'deleteMany':
        await _executeDeleteMany(col, tc, result);
        break;
      case 'replaceOne':
        await _executeReplaceOne(col, tc, result);
        break;
      case 'count':
        _executeCount(col, tc, result);
        break;
      case 'createIndex':
        await _executeCreateIndex(col, tc, result);
        break;
      case 'listIndexes':
        _executeListIndexes(col, tc, result);
        break;
      case 'dropIndex':
        await _executeDropIndex(col, tc, result);
        break;
      case 'findAndModify':
        await _executeFindAndModify(col, tc, result);
        break;
      case 'distinct':
        await _executeDistinct(col, tc, result);
        break;
      case 'aggregate':
        await _executeAggregate(col, tc, result);
        break;
      default:
        throw UnsupportedError('未知方法: ${tc.action.method}');
    }
  }

  /// 执行插入单个文档
  Future<void> _executeInsertOne(Collection col, TestCase tc, TestResult result) async {
    final doc = toBsonDocument(tc.action.doc);
    await col.insert([doc]);
    result.count = 1;
  }

  /// 执行批量插入文档
  Future<void> _executeInsertMany(Collection col, TestCase tc, TestResult result) async {
    final docs = <BsonDocument>[];
    for (final item in tc.action.docs ?? []) {
      docs.add(toBsonDocument(item));
    }
    await col.insert(docs);
    result.count = docs.length;
  }

  /// 执行查询
  Future<void> _executeFind(Collection col, TestCase tc, TestResult result) async {
    BsonDocument? filter;
    if (tc.action.filter != null) {
      filter = toBsonDocument(tc.action.filter);
    }

    QueryOptions? opts;
    if (tc.action.options != null) {
      final optMap = tc.action.options as Map<String, dynamic>;
      opts = QueryOptions(
        sort: toSortMap(optMap['sort']),
        limit: optMap['limit'] as int?,
        skip: optMap['skip'] as int?,
        projection: toSortMap(optMap['projection']),
      );
    }

    List<BsonDocument> docs;
    if (opts != null) {
      docs = await col.findWithOptions(filter, opts);
    } else {
      docs = await col.find(filter);
    }

    result.count = docs.length;
    result.documents = docs.map((d) => documentToMap(d)).toList();
  }

  /// 执行查询单个文档
  Future<void> _executeFindOne(Collection col, TestCase tc, TestResult result) async {
    final filter = toBsonDocument(tc.action.filter);
    final doc = await col.findOne(filter);
    if (doc != null) {
      result.count = 1;
      result.documents = [documentToMap(doc)];
    } else {
      result.count = 0;
    }
  }

  /// 执行更新单个文档
  Future<void> _executeUpdateOne(Collection col, TestCase tc, TestResult result) async {
    final filter = toBsonDocument(tc.action.filter);
    final update = toBsonDocument(tc.action.update);

    final updateResult = await col.update(filter, update);
    result.matchedCount = updateResult.matchedCount;
    result.modifiedCount = updateResult.modifiedCount;
    result.upsertedId = updateResult.upsertedId;
  }

  /// 执行更新多个文档
  Future<void> _executeUpdateMany(Collection col, TestCase tc, TestResult result) async {
    final filter = toBsonDocument(tc.action.filter);
    final update = toBsonDocument(tc.action.update);

    final updateResult = await col.update(filter, update);
    result.matchedCount = updateResult.matchedCount;
    result.modifiedCount = updateResult.modifiedCount;
  }

  /// 执行删除单个文档
  Future<void> _executeDeleteOne(Collection col, TestCase tc, TestResult result) async {
    final filter = toBsonDocument(tc.action.filter);
    final count = await col.deleteOne(filter);
    result.deletedCount = count;
  }

  /// 执行删除多个文档
  Future<void> _executeDeleteMany(Collection col, TestCase tc, TestResult result) async {
    final filter = toBsonDocument(tc.action.filter);
    // 使用 delete 方法代替 deleteMany
    final count = await col.delete(filter);
    result.deletedCount = count;
  }

  /// 执行替换单个文档
  Future<void> _executeReplaceOne(Collection col, TestCase tc, TestResult result) async {
    final filter = toBsonDocument(tc.action.filter);
    final replacement = toBsonDocument(tc.action.doc);

    // 先找到文档
    final existing = await col.findOne(filter);
    if (existing != null) {
      // 保留 _id
      final id = existing['_id'];
      if (id != null) {
        replacement['_id'] = id;
      }
      // 删除旧文档并插入新文档
      await col.deleteOne(filter);
      await col.insert([replacement]);
      result.matchedCount = 1;
      result.modifiedCount = 1;
    } else {
      result.matchedCount = 0;
      result.modifiedCount = 0;
    }
  }

  /// 执行计数
  void _executeCount(Collection col, TestCase tc, TestResult result) {
    result.count = col.count();
  }

  /// 执行创建索引
  Future<void> _executeCreateIndex(Collection col, TestCase tc, TestResult result) async {
    final opts = tc.action.options as Map<String, dynamic>;
    final keys = opts['keys'] as Map<String, dynamic>;
    final indexOpts = opts['options'] as Map<String, dynamic>?;
    final unique = indexOpts?['unique'] as bool? ?? false;
    final name = indexOpts?['name'] as String?;

    final indexName = await col.createIndex(
      keys.map((k, v) => MapEntry(k, (v as num).toInt())),
      name: name,
      unique: unique,
    );
    result.count = 1;
    result.documents = [{'indexName': indexName}];
  }

  /// 执行列出索引
  void _executeListIndexes(Collection col, TestCase tc, TestResult result) {
    final indexes = col.listIndexes();
    result.count = indexes.length;
  }

  /// 执行删除索引
  Future<void> _executeDropIndex(Collection col, TestCase tc, TestResult result) async {
    final opts = tc.action.options as Map<String, dynamic>;
    final name = opts['name'] as String;
    await col.dropIndex(name);
  }

  /// 执行 findAndModify
  Future<void> _executeFindAndModify(Collection col, TestCase tc, TestResult result) async {
    final filter = tc.action.filter != null ? toBsonDocument(tc.action.filter) : null;
    final optMap = tc.action.options as Map<String, dynamic>?;

    final isRemove = optMap?['remove'] as bool? ?? false;
    final returnNew = optMap?['new'] as bool? ?? false;

    if (isRemove) {
      // 删除模式
      final doc = await col.findOne(filter);
      if (doc != null) {
        await col.deleteOne(filter);
        result.count = 1;
        result.documents = [documentToMap(doc)];
      } else {
        result.count = 0;
      }
    } else {
      // 更新模式
      final update = tc.action.update != null ? toBsonDocument(tc.action.update) : null;

      if (returnNew) {
        // 先更新，再返回新文档
        if (update != null) {
          await col.update(filter, update);
        }
        final doc = await col.findOne(filter);
        if (doc != null) {
          result.count = 1;
          result.documents = [documentToMap(doc)];
        } else {
          result.count = 0;
        }
      } else {
        // 先返回旧文档，再更新
        final doc = await col.findOne(filter);
        if (doc != null) {
          result.count = 1;
          result.documents = [documentToMap(doc)];
          if (update != null) {
            await col.update(filter, update);
          }
        } else {
          result.count = 0;
        }
      }
    }
  }

  /// 执行 distinct
  Future<void> _executeDistinct(Collection col, TestCase tc, TestResult result) async {
    final filter = tc.action.filter != null ? toBsonDocument(tc.action.filter) : null;
    final optMap = tc.action.options as Map<String, dynamic>?;
    final field = optMap?['field'] as String? ?? '';

    final values = await col.distinct(field, filter);
    result.count = values.length;
  }

  /// 执行聚合管道（简化实现）
  Future<void> _executeAggregate(Collection col, TestCase tc, TestResult result) async {
    final optMap = tc.action.options as Map<String, dynamic>?;
    final pipeline = optMap?['pipeline'] as List<dynamic>? ?? [];

    // 从集合获取所有文档作为起点
    var docs = await col.find(null);

    // 简化的聚合管道处理
    for (final stage in pipeline) {
      final stageDoc = toBsonDocument(stage);

      if (stageDoc.containsKey('\$match')) {
        // $match 阶段
        final matchFilter = toBsonDocument(stageDoc['\$match']);
        docs = docs.where((doc) => matchesFilter(doc, matchFilter)).toList();
      } else if (stageDoc.containsKey('\$sort')) {
        // $sort 阶段
        final sortSpec = toSortMap(stageDoc['\$sort']);
        if (sortSpec != null) {
          docs = _sortDocuments(docs, sortSpec);
        }
      } else if (stageDoc.containsKey('\$limit')) {
        // $limit 阶段
        final limit = stageDoc['\$limit'] as int;
        if (docs.length > limit) {
          docs = docs.sublist(0, limit);
        }
      } else if (stageDoc.containsKey('\$skip')) {
        // $skip 阶段
        final skip = stageDoc['\$skip'] as int;
        if (docs.length > skip) {
          docs = docs.sublist(skip);
        } else {
          docs = [];
        }
      } else if (stageDoc.containsKey('\$project')) {
        // $project 阶段
        final projSpec = toSortMap(stageDoc['\$project']);
        if (projSpec != null) {
          docs = _applyProjection(docs, projSpec);
        }
      } else if (stageDoc.containsKey('\$group')) {
        // $group 阶段（简化实现）
        final groupSpec = toBsonDocument(stageDoc['\$group']);
        docs = _executeGroup(docs, groupSpec);
      }
    }

    result.count = docs.length;
    result.documents = docs.map((d) => documentToMap(d)).toList();
  }

  /// 对文档排序
  List<BsonDocument> _sortDocuments(List<BsonDocument> docs, Map<String, int> sortSpec) {
    final result = List<BsonDocument>.from(docs);
    result.sort((a, b) {
      for (final entry in sortSpec.entries) {
        final valA = a[entry.key];
        final valB = b[entry.key];
        final cmp = _compareValues(valA, valB);
        if (cmp != 0) {
          return entry.value < 0 ? -cmp : cmp;
        }
      }
      return 0;
    });
    return result;
  }

  /// 比较两个值
  int _compareValues(dynamic a, dynamic b) {
    if (a == null && b == null) return 0;
    if (a == null) return -1;
    if (b == null) return 1;
    if (a is num && b is num) return a.compareTo(b);
    if (a is String && b is String) return a.compareTo(b);
    return a.toString().compareTo(b.toString());
  }

  /// 应用投影
  List<BsonDocument> _applyProjection(List<BsonDocument> docs, Map<String, int> projection) {
    final result = <BsonDocument>[];

    // 判断是包含还是排除模式
    var includeMode = false;
    for (final entry in projection.entries) {
      if (entry.key == '_id') continue;
      if (entry.value == 1) {
        includeMode = true;
      }
      break;
    }

    for (final doc in docs) {
      if (includeMode) {
        final newDoc = BsonDocument();
        // 默认包含 _id
        var includeId = true;
        if (projection['_id'] == 0) includeId = false;
        if (includeId && doc.containsKey('_id')) {
          newDoc['_id'] = doc['_id'];
        }
        for (final entry in projection.entries) {
          if (entry.key == '_id') continue;
          if (entry.value == 1 && doc.containsKey(entry.key)) {
            newDoc[entry.key] = doc[entry.key];
          }
        }
        result.add(newDoc);
      } else {
        final newDoc = BsonDocument();
        final excludeFields = <String>{};
        for (final entry in projection.entries) {
          if (entry.value == 0) excludeFields.add(entry.key);
        }
        for (final entry in doc.entries) {
          if (!excludeFields.contains(entry.key)) {
            newDoc[entry.key] = entry.value;
          }
        }
        result.add(newDoc);
      }
    }
    return result;
  }

  /// 执行 $group 阶段
  List<BsonDocument> _executeGroup(List<BsonDocument> docs, BsonDocument groupSpec) {
    final idSpec = groupSpec['_id'];
    final groups = <String, List<BsonDocument>>{};

    // 按 _id 分组
    for (final doc in docs) {
      String groupKey;
      if (idSpec == null) {
        groupKey = 'null';
      } else if (idSpec is String && idSpec.startsWith('\$')) {
        final field = idSpec.substring(1);
        groupKey = doc[field]?.toString() ?? 'null';
      } else {
        groupKey = idSpec.toString();
      }

      groups.putIfAbsent(groupKey, () => []).add(doc);
    }

    // 构建结果
    final result = <BsonDocument>[];
    for (final entry in groups.entries) {
      final groupDoc = BsonDocument();
      groupDoc['_id'] = entry.key == 'null' ? null : entry.key;

      // 处理聚合操作符
      for (final specEntry in groupSpec.entries) {
        if (specEntry.key == '_id') continue;

        final accSpec = specEntry.value;
        if (accSpec is BsonDocument) {
          if (accSpec.containsKey('\$sum')) {
            final sumSpec = accSpec['\$sum'];
            if (sumSpec == 1) {
              groupDoc[specEntry.key] = entry.value.length;
            } else if (sumSpec is String && sumSpec.startsWith('\$')) {
              final field = sumSpec.substring(1);
              num sum = 0;
              for (final d in entry.value) {
                final val = d[field];
                if (val is num) sum += val;
              }
              groupDoc[specEntry.key] = sum;
            }
          } else if (accSpec.containsKey('\$avg')) {
            final avgSpec = accSpec['\$avg'];
            if (avgSpec is String && avgSpec.startsWith('\$')) {
              final field = avgSpec.substring(1);
              num sum = 0;
              int count = 0;
              for (final d in entry.value) {
                final val = d[field];
                if (val is num) {
                  sum += val;
                  count++;
                }
              }
              groupDoc[specEntry.key] = count > 0 ? sum / count : 0;
            }
          }
        }
      }

      result.add(groupDoc);
    }

    return result;
  }
}
