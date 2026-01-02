// Created by Yanjunhui

import 'dart:convert';
import 'dart:io';

import 'package:args/args.dart';
import 'package:monolite_test_runner/types.dart';
import 'package:monolite_test_runner/api_runner.dart';

/// 命令行参数
late String mode;
late String monoDBPath;
late String testCasesPath;
late String outputPath;

void main(List<String> args) async {
  // 解析命令行参数
  final parser = ArgParser()
    ..addOption('mode', abbr: 'm', defaultsTo: 'api', help: '测试模式: api 或 wire')
    ..addOption('monodb', defaultsTo: '../../testdata/fixtures/test.monodb', help: 'MonoLite 数据库文件')
    ..addOption('testcases', defaultsTo: '../../testdata/fixtures/testcases.json', help: '测试用例文件')
    ..addOption('output', abbr: 'o', defaultsTo: '../../reports/dart_results.json', help: '结果输出文件');

  final results = parser.parse(args);

  mode = results['mode'] as String;
  monoDBPath = results['monodb'] as String;
  testCasesPath = results['testcases'] as String;
  outputPath = results['output'] as String;

  print('=== Dart 测试运行器 ($mode 模式) ===');

  // 加载测试用例
  final suite = await loadTestCases(testCasesPath);
  print('加载了 ${suite.tests.length} 个测试用例');

  List<TestResult> testResults;
  int passed, failed;

  switch (mode) {
    case 'api':
      (testResults, passed, failed) = await runAPITests(suite);
      break;
    case 'wire':
      print('警告: Wire 模式在 Dart 版本中尚未实现，将跳过所有测试');
      testResults = suite.tests.map((tc) => TestResult(
        testName: tc.name,
        language: 'dart',
        mode: 'wire',
        success: false,
        error: 'Wire Protocol not implemented in Dart version',
      )).toList();
      passed = 0;
      failed = testResults.length;
      break;
    default:
      print('未知模式: $mode');
      exit(1);
  }

  // 保存结果
  final resultsFile = ResultsFile(
    language: 'dart',
    mode: mode,
    results: testResults,
    summary: Summary(
      total: testResults.length,
      passed: passed,
      failed: failed,
    ),
  );

  await saveResults(outputPath, resultsFile);

  print('=== 测试完成 ===');
  print('通过: $passed, 失败: $failed, 总计: ${testResults.length}');
  print('结果已保存到: $outputPath');
}

/// 加载测试用例
Future<TestSuite> loadTestCases(String path) async {
  final file = File(path);
  if (!await file.exists()) {
    throw FileSystemException('测试用例文件不存在', path);
  }
  final content = await file.readAsString();
  final json = jsonDecode(content) as Map<String, dynamic>;
  return TestSuite.fromJson(json);
}

/// 运行 API 模式测试
Future<(List<TestResult>, int, int)> runAPITests(TestSuite suite) async {
  // 删除现有数据库文件，确保从干净状态开始
  final dbFile = File(monoDBPath);
  if (await dbFile.exists()) {
    await dbFile.delete();
  }
  final walFile = File('$monoDBPath.wal');
  if (await walFile.exists()) {
    await walFile.delete();
  }

  final runner = await APIRunner.create(monoDBPath);

  final results = <TestResult>[];
  var passed = 0;
  var failed = 0;

  for (var i = 0; i < suite.tests.length; i++) {
    final tc = suite.tests[i];
    print('[${i + 1}/${suite.tests.length}] 测试: ${tc.name}');

    try {
      final result = await runner.runTest(tc);
      results.add(result);

      if (result.success) {
        passed++;
        print('  ✓ 通过 (${result.durationMs}ms)');
      } else {
        failed++;
        print('  ✗ 失败: ${result.error} (${result.durationMs}ms)');
      }
    } catch (e, st) {
      failed++;
      final result = TestResult(
        testName: tc.name,
        language: 'dart',
        mode: 'api',
        success: false,
        error: '$e\n$st',
      );
      results.add(result);
      print('  ✗ 异常: $e');
    }
  }

  await runner.close();

  return (results, passed, failed);
}

/// 保存测试结果
Future<void> saveResults(String path, ResultsFile results) async {
  final file = File(path);
  await file.parent.create(recursive: true);
  final json = const JsonEncoder.withIndent('  ').convert(results.toJson());
  await file.writeAsString(json);
}
