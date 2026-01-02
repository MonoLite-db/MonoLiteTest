// Created by Yanjunhui

import 'package:monolite/monolite.dart';

/// 测试用例定义
/// EN: TestCase defines a test case structure.
class TestCase {
  final String name; // 测试名称
  final String category; // 分类
  final String operation; // 操作类型
  final String collection; // 集合名称
  final String description; // 描述
  final List<SetupStep> setup; // 前置步骤
  final TestAction action; // 测试动作
  final Expected expected; // 预期结果

  TestCase({
    required this.name,
    required this.category,
    required this.operation,
    required this.collection,
    required this.description,
    required this.setup,
    required this.action,
    required this.expected,
  });

  factory TestCase.fromJson(Map<String, dynamic> json) {
    return TestCase(
      name: json['name'] as String? ?? '',
      category: json['category'] as String? ?? '',
      operation: json['operation'] as String? ?? '',
      collection: json['collection'] as String? ?? '',
      description: json['description'] as String? ?? '',
      setup: (json['setup'] as List<dynamic>?)
              ?.map((e) => SetupStep.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      action: TestAction.fromJson(json['action'] as Map<String, dynamic>? ?? {}),
      expected: Expected.fromJson(json['expected'] as Map<String, dynamic>? ?? {}),
    );
  }
}

/// 前置步骤
/// EN: SetupStep defines a setup step before test execution.
class SetupStep {
  final String operation; // 操作类型
  final dynamic data; // 操作数据

  SetupStep({
    required this.operation,
    this.data,
  });

  factory SetupStep.fromJson(Map<String, dynamic> json) {
    return SetupStep(
      operation: json['operation'] as String? ?? '',
      data: json['data'],
    );
  }
}

/// 测试动作
/// EN: TestAction defines the action to be performed in a test.
class TestAction {
  final String method; // 方法名
  final dynamic filter; // 查询条件
  final dynamic update; // 更新内容
  final dynamic doc; // 文档
  final List<dynamic>? docs; // 文档列表
  final dynamic options; // 选项

  TestAction({
    required this.method,
    this.filter,
    this.update,
    this.doc,
    this.docs,
    this.options,
  });

  factory TestAction.fromJson(Map<String, dynamic> json) {
    return TestAction(
      method: json['method'] as String? ?? '',
      filter: json['filter'],
      update: json['update'],
      doc: json['doc'],
      docs: json['docs'] as List<dynamic>?,
      options: json['options'],
    );
  }
}

/// 预期结果
/// EN: Expected defines the expected result of a test.
class Expected {
  final int? count; // 预期数量
  final List<dynamic>? documents; // 预期文档
  final int? matchedCount; // 匹配数量
  final int? modifiedCount; // 修改数量
  final int? deletedCount; // 删除数量
  final dynamic upsertedId; // Upsert ID
  final String? error; // 预期错误
  final String? indexName; // 索引名称

  Expected({
    this.count,
    this.documents,
    this.matchedCount,
    this.modifiedCount,
    this.deletedCount,
    this.upsertedId,
    this.error,
    this.indexName,
  });

  factory Expected.fromJson(Map<String, dynamic> json) {
    return Expected(
      count: json['count'] as int?,
      documents: json['documents'] as List<dynamic>?,
      matchedCount: json['matched_count'] as int?,
      modifiedCount: json['modified_count'] as int?,
      deletedCount: json['deleted_count'] as int?,
      upsertedId: json['upserted_id'],
      error: json['error'] as String?,
      indexName: json['index_name'] as String?,
    );
  }
}

/// 测试结果
/// EN: TestResult defines the result of a test execution.
class TestResult {
  String testName; // 测试名称
  String language; // 语言
  String mode; // 模式
  bool success; // 是否成功
  String? error; // 错误信息
  int durationMs; // 耗时（毫秒）
  List<Map<String, dynamic>>? documents; // 返回的文档
  int count; // 数量
  int matchedCount; // 匹配数量
  int modifiedCount; // 修改数量
  int deletedCount; // 删除数量
  dynamic upsertedId; // Upsert ID

  TestResult({
    required this.testName,
    this.language = 'dart',
    this.mode = 'api',
    this.success = false,
    this.error,
    this.durationMs = 0,
    this.documents,
    this.count = 0,
    this.matchedCount = 0,
    this.modifiedCount = 0,
    this.deletedCount = 0,
    this.upsertedId,
  });

  Map<String, dynamic> toJson() {
    final json = <String, dynamic>{
      'test_name': testName,
      'language': language,
      'mode': mode,
      'success': success,
      'duration_ms': durationMs,
    };
    if (error != null) json['error'] = error;
    if (documents != null) json['documents'] = documents;
    if (count > 0) json['count'] = count;
    if (matchedCount > 0) json['matched_count'] = matchedCount;
    if (modifiedCount > 0) json['modified_count'] = modifiedCount;
    if (deletedCount > 0) json['deleted_count'] = deletedCount;
    if (upsertedId != null) json['upserted_id'] = upsertedId;
    return json;
  }
}

/// 测试套件
/// EN: TestSuite defines a collection of test cases.
class TestSuite {
  final String version; // 版本
  final String generated; // 生成时间
  final List<TestCase> tests; // 测试用例列表

  TestSuite({
    required this.version,
    required this.generated,
    required this.tests,
  });

  factory TestSuite.fromJson(Map<String, dynamic> json) {
    return TestSuite(
      version: json['version'] as String? ?? '',
      generated: json['generated'] as String? ?? '',
      tests: (json['tests'] as List<dynamic>?)
              ?.map((e) => TestCase.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

/// 结果文件
/// EN: ResultsFile defines the structure of a results file.
class ResultsFile {
  final String language; // 语言
  final String mode; // 模式
  final List<TestResult> results; // 结果列表
  final Summary summary; // 摘要

  ResultsFile({
    required this.language,
    required this.mode,
    required this.results,
    required this.summary,
  });

  Map<String, dynamic> toJson() {
    return {
      'language': language,
      'mode': mode,
      'results': results.map((r) => r.toJson()).toList(),
      'summary': summary.toJson(),
    };
  }
}

/// 摘要
/// EN: Summary defines the summary of test results.
class Summary {
  final int total; // 总数
  final int passed; // 通过数
  final int failed; // 失败数
  final int skipped; // 跳过数

  Summary({
    required this.total,
    required this.passed,
    required this.failed,
    this.skipped = 0,
  });

  Map<String, dynamic> toJson() {
    return {
      'total': total,
      'passed': passed,
      'failed': failed,
      'skipped': skipped,
    };
  }
}

/// 将 dynamic 转换为 BsonDocument
BsonDocument toBsonDocument(dynamic value) {
  if (value == null) return BsonDocument();

  if (value is BsonDocument) return value;

  if (value is Map) {
    final doc = BsonDocument();
    for (final entry in value.entries) {
      doc[entry.key.toString()] = convertValue(entry.value);
    }
    return doc;
  }

  return BsonDocument();
}

/// 将 dynamic 转换为 BsonArray
BsonArray toBsonArray(dynamic value) {
  if (value == null) return BsonArray();

  if (value is BsonArray) return value;

  if (value is List) {
    final arr = BsonArray();
    for (final item in value) {
      arr.add(convertValue(item));
    }
    return arr;
  }

  return BsonArray();
}

/// 转换值类型
dynamic convertValue(dynamic value) {
  if (value == null) return null;

  if (value is Map) {
    return toBsonDocument(value);
  }

  if (value is List) {
    return toBsonArray(value);
  }

  // 基本类型直接返回
  return value;
}

/// 将 BsonDocument 转换为 Map
Map<String, dynamic> documentToMap(BsonDocument doc) {
  final map = <String, dynamic>{};
  for (final key in doc.keys) {
    final value = doc[key];
    if (value is BsonDocument) {
      map[key] = documentToMap(value);
    } else if (value is BsonArray) {
      map[key] = arrayToList(value);
    } else if (value is ObjectId) {
      map[key] = value.toHex();
    } else {
      map[key] = value;
    }
  }
  return map;
}

/// 将 BsonArray 转换为 List
List<dynamic> arrayToList(BsonArray arr) {
  final list = <dynamic>[];
  for (var i = 0; i < arr.length; i++) {
    final value = arr[i];
    if (value is BsonDocument) {
      list.add(documentToMap(value));
    } else if (value is BsonArray) {
      list.add(arrayToList(value));
    } else if (value is ObjectId) {
      list.add(value.toHex());
    } else {
      list.add(value);
    }
  }
  return list;
}

/// 将 dynamic 转换为 Map<String, int>（用于 sort/projection）
Map<String, int>? toSortMap(dynamic value) {
  if (value == null) return null;

  if (value is Map) {
    final result = <String, int>{};
    for (final entry in value.entries) {
      final key = entry.key.toString();
      final val = entry.value;
      if (val is num) {
        result[key] = val.toInt();
      }
    }
    return result.isEmpty ? null : result;
  }

  return null;
}
