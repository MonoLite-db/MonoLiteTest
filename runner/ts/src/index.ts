// Created by Yanjunhui

import { Command } from 'commander';
import * as fs from 'fs';
import * as path from 'path';
import { APIRunner } from './apiRunner';
import { WireRunner } from './wireRunner';
import { TestSuite, TestResult, ResultsFile } from './types';

const program = new Command();

program
  .option('-m, --mode <mode>', '测试模式: api 或 wire', 'api')
  .option(
    '-d, --monodb <path>',
    'MonoLite 数据库文件',
    '../../testdata/fixtures/test_ts.monodb'
  )
  .option(
    '-t, --testcases <path>',
    '测试用例文件',
    '../../testdata/fixtures/testcases.json'
  )
  .option('-o, --output <path>', '结果输出文件', '../../reports/ts_results.json')
  .option('-p, --port <port>', 'Wire Protocol 端口', '27019');

program.parse();

const opts = program.opts();

async function main(): Promise<void> {
  console.log(`=== TypeScript 测试运行器 (${opts.mode} 模式) ===`);

  // 解析路径
  const monodbPath = path.resolve(__dirname, '..', opts.monodb);
  const testcasesPath = path.resolve(__dirname, '..', opts.testcases);
  const outputPath = path.resolve(__dirname, '..', opts.output);

  // 删除现有数据库文件（每次运行创建新的）
  if (fs.existsSync(monodbPath)) {
    fs.unlinkSync(monodbPath);
    console.log(`已删除现有数据库: ${monodbPath}`);
  }

  // 加载测试用例
  const suiteData = fs.readFileSync(testcasesPath, 'utf-8');
  const suite: TestSuite = JSON.parse(suiteData);
  console.log(`加载了 ${suite.tests.length} 个测试用例`);

  let results: TestResult[];
  let passed = 0;
  let failed = 0;

  if (opts.mode === 'api') {
    const runner = new APIRunner(monodbPath);
    await runner.open();

    results = [];
    for (let i = 0; i < suite.tests.length; i++) {
      const tc = suite.tests[i];
      console.log(`[${i + 1}/${suite.tests.length}] 测试: ${tc.name}`);

      const result = await runner.runTest(tc);
      results.push(result);

      if (result.success) {
        passed++;
        console.log(`  ✓ 通过 (${result.duration_ms}ms)`);
      } else {
        failed++;
        console.log(`  ✗ 失败: ${result.error} (${result.duration_ms}ms)`);
      }
    }

    await runner.close();
  } else if (opts.mode === 'wire') {
    const port = parseInt(opts.port, 10);
    const runner = new WireRunner(monodbPath, port);
    await runner.open();

    results = [];
    for (let i = 0; i < suite.tests.length; i++) {
      const tc = suite.tests[i];
      console.log(`[${i + 1}/${suite.tests.length}] 测试: ${tc.name}`);

      const result = await runner.runTest(tc);
      results.push(result);

      if (result.success) {
        passed++;
        console.log(`  ✓ 通过 (${result.duration_ms}ms)`);
      } else {
        failed++;
        console.log(`  ✗ 失败: ${result.error} (${result.duration_ms}ms)`);
      }
    }

    await runner.close();
  } else {
    console.error(`未知模式: ${opts.mode}`);
    process.exit(1);
  }

  // 保存结果
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
  console.log(`通过: ${passed}, 失败: ${failed}, 总计: ${results.length}`);
  console.log(`结果已保存到: ${outputPath}`);
}

main().catch((err) => {
  console.error('运行失败:', err);
  process.exit(1);
});
