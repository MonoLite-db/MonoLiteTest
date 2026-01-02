// Created by Yanjunhui

import 'types.dart';

/// Wire Protocol 模式测试运行器（占位符）
/// EN: WireRunner tests through Wire Protocol (placeholder).
///
/// 注意：Dart 版本的 MonoLite 尚未实现 Wire Protocol。
/// Note: The Dart version of MonoLite does not yet implement Wire Protocol.
class WireRunner {
  WireRunner._();

  /// 创建 Wire 运行器
  /// EN: Creates a Wire runner.
  static Future<WireRunner> create(String dbPath, int port) async {
    throw UnsupportedError(
      'Wire Protocol 尚未在 Dart 版本中实现。'
      'Wire Protocol is not yet implemented in the Dart version.',
    );
  }

  /// 关闭连接
  /// EN: Closes all connections.
  Future<void> close() async {
    // Not implemented
  }

  /// 运行单个测试
  /// EN: Runs a single test case.
  Future<TestResult> runTest(TestCase tc) async {
    return TestResult(
      testName: tc.name,
      language: 'dart',
      mode: 'wire',
      success: false,
      error: 'Wire Protocol not implemented',
    );
  }
}
