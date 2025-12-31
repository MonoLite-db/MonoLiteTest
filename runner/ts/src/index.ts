// Created by Yanjunhui

import { Command } from 'commander';
import * as fs from 'fs';
import * as path from 'path';
import { APIRunner } from './apiRunner';
import { WireRunner } from './wireRunner';
import { TestSuite, TestResult, ResultsFile } from './types';

// 创建命令行解析器 // EN: Create command line parser
const program = new Command();

program
  .option('-m, --mode <mode>', '测试模式: api 或 wire // EN: Test mode: api or wire', 'api')
  .option(
    '-d, --monodb <path>',
    'MonoLite 数据库文件 // EN: MonoLite database file',
    '../../testdata/fixtures/test_ts.monodb'
  )
  .option(
    '-t, --testcases <path>',
    '测试用例文件 // EN: Test cases file',
    '../../testdata/fixtures/testcases.json'
  )
  .option('-o, --output <path>', '结果输出文件 // EN: Result output file', '../../reports/ts_results.json')
  .option('-p, --port <port>', 'Wire Protocol 端口 // EN: Wire Protocol port', '27019');

program.parse();

const opts = program.opts();

/**
 * 主函数，执行测试运行器的入口点
 * // EN: Main function, entry point for test runner execution
 */
async function main(): Promise<void> {
  console.log(`=== TypeScript 测试运行器 (${opts.mode} 模式) ===`);
  // // EN: === TypeScript Test Runner ({mode} mode) ===

  // 解析路径 // EN: Resolve paths
  const monodbPath = path.resolve(__dirname, '..', opts.monodb);
  const testcasesPath = path.resolve(__dirname, '..', opts.testcases);
  const outputPath = path.resolve(__dirname, '..', opts.output);

  // 删除现有数据库文件（每次运行创建新的）
  // // EN: Delete existing database file (create new one for each run)
  if (fs.existsSync(monodbPath)) {
    fs.unlinkSync(monodbPath);
    console.log(`已删除现有数据库: ${monodbPath}`);
    // // EN: Deleted existing database: {path}
  }

  // 加载测试用例 // EN: Load test cases
  const suiteData = fs.readFileSync(testcasesPath, 'utf-8');
  const suite: TestSuite = JSON.parse(suiteData);
  console.log(`加载了 ${suite.tests.length} 个测试用例`);
  // // EN: Loaded {count} test cases

  let results: TestResult[];
  let passed = 0;
  let failed = 0;

  if (opts.mode === 'api') {
    // API 模式：直接调用 MonoLite API
    // // EN: API mode: call MonoLite API directly
    const runner = new APIRunner(monodbPath);
    await runner.open();

    results = [];
    for (let i = 0; i < suite.tests.length; i++) {
      const tc = suite.tests[i];
      console.log(`[${i + 1}/${suite.tests.length}] 测试: ${tc.name}`);
      // // EN: [{current}/{total}] Test: {name}

      const result = await runner.runTest(tc);
      results.push(result);

      if (result.success) {
        passed++;
        console.log(`  ✓ 通过 (${result.duration_ms}ms)`);
        // // EN: ✓ Passed ({duration}ms)
      } else {
        failed++;
        console.log(`  ✗ 失败: ${result.error} (${result.duration_ms}ms)`);
        // // EN: ✗ Failed: {error} ({duration}ms)
      }
    }

    await runner.close();
  } else if (opts.mode === 'wire') {
    // Wire 模式：通过 MongoDB Wire Protocol 连接
    // // EN: Wire mode: connect via MongoDB Wire Protocol
    const port = parseInt(opts.port, 10);
    const runner = new WireRunner(monodbPath, port);
    await runner.open();

    results = [];
    for (let i = 0; i < suite.tests.length; i++) {
      const tc = suite.tests[i];
      console.log(`[${i + 1}/${suite.tests.length}] 测试: ${tc.name}`);
      // // EN: [{current}/{total}] Test: {name}

      const result = await runner.runTest(tc);
      results.push(result);

      if (result.success) {
        passed++;
        console.log(`  ✓ 通过 (${result.duration_ms}ms)`);
        // // EN: ✓ Passed ({duration}ms)
      } else {
        failed++;
        console.log(`  ✗ 失败: ${result.error} (${result.duration_ms}ms)`);
        // // EN: ✗ Failed: {error} ({duration}ms)
      }
    }

    await runner.close();
  } else {
    // 未知模式 // EN: Unknown mode
    console.error(`未知模式: ${opts.mode}`);
    // // EN: Unknown mode: {mode}
    process.exit(1);
  }

  // 保存结果 // EN: Save results
  const resultsFile: ResultsFile = {
    language: 'typescript',
    mode: opts.mode,
    results,
    summary: {
      total: results.length,
      passed,
      failed,
      skipped: 0,
    },
  };

  fs.writeFileSync(outputPath, JSON.stringify(resultsFile, null, 2));

  console.log('=== 测试完成 ===');
  // // EN: === Tests completed ===
  console.log(`通过: ${passed}, 失败: ${failed}, 总计: ${results.length}`);
  // // EN: Passed: {passed}, Failed: {failed}, Total: {total}
  console.log(`结果已保存到: ${outputPath}`);
  // // EN: Results saved to: {path}
}

// 执行主函数 // EN: Execute main function
main().catch((err) => {
  console.error('运行失败:', err);
  // // EN: Run failed: {error}
  process.exit(1);
});
