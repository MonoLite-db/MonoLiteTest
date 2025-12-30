// Created by Yanjunhui

import Foundation
import ArgumentParser
import MonoLiteSwift

struct Runner: ParsableCommand {
    static var configuration = CommandConfiguration(
        commandName: "Runner",
        abstract: "MonoLite Swift 测试运行器"
    )

    @Option(name: .shortAndLong, help: "测试模式: api 或 wire")
    var mode: String = "api"

    @Option(name: [.customShort("d"), .customLong("monodb")], help: "MonoLite 数据库文件")
    var monodb: String = "../../testdata/fixtures/test_swift.monodb"

    @Option(name: .shortAndLong, help: "测试用例文件")
    var testcases: String = "../../testdata/fixtures/testcases.json"

    @Option(name: .shortAndLong, help: "结果输出文件")
    var output: String = "../../reports/swift_results.json"

    @Option(name: .shortAndLong, help: "Wire Protocol 端口")
    var port: Int = 27020

    func run() throws {
        // 使用 Task 和 semaphore 来同步运行异步代码
        let semaphore = DispatchSemaphore(value: 0)
        var runError: Error?

        // 捕获配置参数
        let mode = self.mode
        let monodb = self.monodb
        let testcases = self.testcases
        let output = self.output
        let port = self.port

        Task {
            do {
                try await Self.runAsync(mode: mode, monodb: monodb, testcases: testcases, output: output, port: port)
            } catch {
                runError = error
            }
            semaphore.signal()
        }

        semaphore.wait()
        if let error = runError {
            throw error
        }
    }

    static func runAsync(mode: String, monodb: String, testcases: String, output: String, port: Int) async throws {
        print("=== Swift 测试运行器 (\(mode) 模式) ===")

        // 解析路径
        let currentDir = FileManager.default.currentDirectoryPath
        let monodbPath = resolvePath(monodb, from: currentDir)
        let testcasesPath = resolvePath(testcases, from: currentDir)
        let outputPath = resolvePath(output, from: currentDir)

        // 删除现有数据库文件（每次运行创建新的）
        if FileManager.default.fileExists(atPath: monodbPath) {
            try FileManager.default.removeItem(atPath: monodbPath)
            print("已删除现有数据库: \(monodbPath)")
        }

        // 加载测试用例
        guard FileManager.default.fileExists(atPath: testcasesPath) else {
            print("错误: 测试用例文件不存在: \(testcasesPath)")
            return
        }

        let testcasesData = try Data(contentsOf: URL(fileURLWithPath: testcasesPath))
        let decoder = JSONDecoder()
        let suite = try decoder.decode(TestSuite.self, from: testcasesData)
        print("加载了 \(suite.tests.count) 个测试用例")

        var results: [TestResult] = []
        var passed = 0
        var failed = 0

        if mode == "api" {
            let runner = APIRunner(dbPath: monodbPath)
            try await runner.open()

            for (index, tc) in suite.tests.enumerated() {
                print("[\(index + 1)/\(suite.tests.count)] 测试: \(tc.name)")

                let result = await runner.runTest(tc)
                results.append(result)

                if result.success {
                    passed += 1
                    print("  ✓ 通过 (\(result.duration_ms)ms)")
                } else {
                    failed += 1
                    print("  ✗ 失败: \(result.error ?? "unknown") (\(result.duration_ms)ms)")
                }
            }

            try await runner.close()
        } else if mode == "wire" {
            let runner = WireRunner(dbPath: monodbPath, port: port)
            try await runner.open()

            for (index, tc) in suite.tests.enumerated() {
                print("[\(index + 1)/\(suite.tests.count)] 测试: \(tc.name)")

                let result = await runner.runTest(tc)
                results.append(result)

                if result.success {
                    passed += 1
                    print("  ✓ 通过 (\(result.duration_ms)ms)")
                } else {
                    failed += 1
                    print("  ✗ 失败: \(result.error ?? "unknown") (\(result.duration_ms)ms)")
                }
            }

            try await runner.close()
        } else {
            print("未知模式: \(mode)")
            return
        }

        // 保存结果
        let resultsFile = ResultsFile(
            language: "swift",
            mode: mode,
            results: results,
            summary: Summary(
                total: results.count,
                passed: passed,
                failed: failed,
                skipped: 0
            )
        )

        let encoder = JSONEncoder()
        encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
        let outputData = try encoder.encode(resultsFile)

        // 确保输出目录存在
        let outputDir = (outputPath as NSString).deletingLastPathComponent
        try FileManager.default.createDirectory(atPath: outputDir, withIntermediateDirectories: true)

        try outputData.write(to: URL(fileURLWithPath: outputPath))

        print("=== 测试完成 ===")
        print("通过: \(passed), 失败: \(failed), 总计: \(results.count)")
        print("结果已保存到: \(outputPath)")
    }

    private static func resolvePath(_ path: String, from base: String) -> String {
        if path.hasPrefix("/") {
            return path
        }
        return (base as NSString).appendingPathComponent(path)
    }
}

Runner.main()
