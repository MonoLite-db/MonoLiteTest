// Created by Yanjunhui

import Foundation
import ArgumentParser
import MonoLiteSwift

/// MonoLite Swift 测试运行器命令行工具
/// EN: MonoLite Swift test runner command-line tool
struct Runner: ParsableCommand {
    static var configuration = CommandConfiguration(
        commandName: "Runner",
        abstract: "MonoLite Swift 测试运行器" // EN: MonoLite Swift test runner
    )

    /// 测试模式：api 或 wire
    /// EN: Test mode: api or wire
    @Option(name: .shortAndLong, help: "测试模式: api 或 wire") // EN: Test mode: api or wire
    var mode: String = "api"

    /// MonoLite 数据库文件路径
    /// EN: MonoLite database file path
    @Option(name: [.customShort("d"), .customLong("monodb")], help: "MonoLite 数据库文件") // EN: MonoLite database file
    var monodb: String = "../../testdata/fixtures/test_swift.monodb"

    /// 测试用例 JSON 文件路径
    /// EN: Test cases JSON file path
    @Option(name: .shortAndLong, help: "测试用例文件") // EN: Test cases file
    var testcases: String = "../../testdata/fixtures/testcases.json"

    /// 结果输出 JSON 文件路径
    /// EN: Results output JSON file path
    @Option(name: .shortAndLong, help: "结果输出文件") // EN: Results output file
    var output: String = "../../reports/swift_results.json"

    /// Wire Protocol 服务端口
    /// EN: Wire Protocol server port
    @Option(name: .shortAndLong, help: "Wire Protocol 端口") // EN: Wire Protocol port
    var port: Int = 27020

    /// 运行命令主入口
    /// EN: Main entry point for running the command
    func run() throws {
        // 使用 Task 和 semaphore 来同步运行异步代码
        // EN: Use Task and semaphore to synchronously run async code
        let semaphore = DispatchSemaphore(value: 0)
        var runError: Error?

        // 捕获配置参数 // EN: Capture configuration parameters
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

    /// 异步运行测试
    /// EN: Run tests asynchronously
    /// - Parameters:
    ///   - mode: 测试模式 / Test mode
    ///   - monodb: 数据库文件路径 / Database file path
    ///   - testcases: 测试用例文件路径 / Test cases file path
    ///   - output: 输出文件路径 / Output file path
    ///   - port: Wire Protocol 端口 / Wire Protocol port
    static func runAsync(mode: String, monodb: String, testcases: String, output: String, port: Int) async throws {
        print("=== Swift 测试运行器 (\(mode) 模式) ===") // EN: Swift test runner (mode) mode

        // 解析路径 // EN: Resolve paths
        let currentDir = FileManager.default.currentDirectoryPath
        let monodbPath = resolvePath(monodb, from: currentDir)
        let testcasesPath = resolvePath(testcases, from: currentDir)
        let outputPath = resolvePath(output, from: currentDir)

        // 删除现有数据库文件（每次运行创建新的）
        // EN: Delete existing database file (create new one each run)
        if FileManager.default.fileExists(atPath: monodbPath) {
            try FileManager.default.removeItem(atPath: monodbPath)
            print("已删除现有数据库: \(monodbPath)") // EN: Deleted existing database
        }

        // 加载测试用例 // EN: Load test cases
        guard FileManager.default.fileExists(atPath: testcasesPath) else {
            print("错误: 测试用例文件不存在: \(testcasesPath)") // EN: Error: Test cases file not found
            return
        }

        let testcasesData = try Data(contentsOf: URL(fileURLWithPath: testcasesPath))
        let decoder = JSONDecoder()
        let suite = try decoder.decode(TestSuite.self, from: testcasesData)
        print("加载了 \(suite.tests.count) 个测试用例") // EN: Loaded N test cases

        var results: [TestResult] = []
        var passed = 0
        var failed = 0

        if mode == "api" {
            // API 模式：直接调用 MonoLiteSwift API
            // EN: API mode: Call MonoLiteSwift API directly
            let runner = APIRunner(dbPath: monodbPath)
            try await runner.open()

            for (index, tc) in suite.tests.enumerated() {
                print("[\(index + 1)/\(suite.tests.count)] 测试: \(tc.name)") // EN: Test

                let result = await runner.runTest(tc)
                results.append(result)

                if result.success {
                    passed += 1
                    print("  ✓ 通过 (\(result.duration_ms)ms)") // EN: Passed
                } else {
                    failed += 1
                    print("  ✗ 失败: \(result.error ?? "unknown") (\(result.duration_ms)ms)") // EN: Failed
                }
            }

            try await runner.close()
        } else if mode == "wire" {
            // Wire 模式：通过 MongoDB Wire Protocol 执行测试
            // EN: Wire mode: Execute tests via MongoDB Wire Protocol
            let runner = WireRunner(dbPath: monodbPath, port: port)
            try await runner.open()

            for (index, tc) in suite.tests.enumerated() {
                print("[\(index + 1)/\(suite.tests.count)] 测试: \(tc.name)") // EN: Test

                let result = await runner.runTest(tc)
                results.append(result)

                if result.success {
                    passed += 1
                    print("  ✓ 通过 (\(result.duration_ms)ms)") // EN: Passed
                } else {
                    failed += 1
                    print("  ✗ 失败: \(result.error ?? "unknown") (\(result.duration_ms)ms)") // EN: Failed
                }
            }

            try await runner.close()
        } else {
            print("未知模式: \(mode)") // EN: Unknown mode
            return
        }

        // 保存结果 // EN: Save results
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

        // 确保输出目录存在 // EN: Ensure output directory exists
        let outputDir = (outputPath as NSString).deletingLastPathComponent
        try FileManager.default.createDirectory(atPath: outputDir, withIntermediateDirectories: true)

        try outputData.write(to: URL(fileURLWithPath: outputPath))

        print("=== 测试完成 ===") // EN: Tests completed
        print("通过: \(passed), 失败: \(failed), 总计: \(results.count)") // EN: Passed, Failed, Total
        print("结果已保存到: \(outputPath)") // EN: Results saved to
    }

    /// 解析相对路径为绝对路径
    /// EN: Resolve relative path to absolute path
    /// - Parameters:
    ///   - path: 输入路径 / Input path
    ///   - base: 基准目录 / Base directory
    /// - Returns: 绝对路径 / Absolute path
    private static func resolvePath(_ path: String, from base: String) -> String {
        if path.hasPrefix("/") {
            return path
        }
        return (base as NSString).appendingPathComponent(path)
    }
}

// 启动命令行工具 // EN: Start command-line tool
Runner.main()
