// Created by Yanjunhui

import Foundation

// MARK: - 测试用例定义 / Test Case Definitions

/// 测试用例结构体，表示一个完整的测试场景
/// EN: Test case struct representing a complete test scenario
struct TestCase: Codable {
    /// 测试名称
    /// EN: Test name
    let name: String
    /// 测试分类
    /// EN: Test category
    let category: String
    /// 操作类型
    /// EN: Operation type
    let operation: String
    /// 集合名称
    /// EN: Collection name
    let collection: String
    /// 测试描述
    /// EN: Test description
    let description: String
    /// 前置步骤（可选）
    /// EN: Setup steps (optional)
    let setup: [SetupStep]?
    /// 测试动作
    /// EN: Test action
    let action: TestAction
    /// 预期结果
    /// EN: Expected result
    let expected: Expected
}

/// 前置步骤结构体，定义测试前的准备操作
/// EN: Setup step struct defining pre-test preparation operations
struct SetupStep: Codable {
    /// 操作类型
    /// EN: Operation type
    let operation: String
    /// 操作数据
    /// EN: Operation data
    let data: AnyCodable
}

/// 测试动作结构体，定义要执行的测试操作
/// EN: Test action struct defining the test operation to execute
struct TestAction: Codable {
    /// 方法名称
    /// EN: Method name
    let method: String
    /// 查询过滤条件（可选）
    /// EN: Query filter condition (optional)
    let filter: AnyCodable?
    /// 更新操作（可选）
    /// EN: Update operation (optional)
    let update: AnyCodable?
    /// 单个文档（可选）
    /// EN: Single document (optional)
    let doc: AnyCodable?
    /// 文档数组（可选）
    /// EN: Document array (optional)
    let docs: [AnyCodable]?
    /// 操作选项（可选）
    /// EN: Operation options (optional)
    let options: AnyCodable?
}

/// 预期结果结构体，定义测试的期望输出
/// EN: Expected result struct defining the expected test output
struct Expected: Codable {
    /// 预期文档数量（可选）
    /// EN: Expected document count (optional)
    let count: Int?
    /// 预期文档列表（可选）
    /// EN: Expected document list (optional)
    let documents: [AnyCodable]?
    /// 预期匹配数量（可选）
    /// EN: Expected matched count (optional)
    let matched_count: Int?
    /// 预期修改数量（可选）
    /// EN: Expected modified count (optional)
    let modified_count: Int?
    /// 预期删除数量（可选）
    /// EN: Expected deleted count (optional)
    let deleted_count: Int?
    /// 预期插入的文档ID（可选）
    /// EN: Expected upserted document ID (optional)
    let upserted_id: AnyCodable?
    /// 预期错误信息（可选）
    /// EN: Expected error message (optional)
    let error: String?
    /// 预期索引名称（可选）
    /// EN: Expected index name (optional)
    let index_name: String?
}

/// 测试套件结构体，包含版本信息和所有测试用例
/// EN: Test suite struct containing version info and all test cases
struct TestSuite: Codable {
    /// 测试套件版本
    /// EN: Test suite version
    let version: String
    /// 生成时间
    /// EN: Generation time
    let generated: String
    /// 测试用例列表
    /// EN: Test case list
    let tests: [TestCase]
}

// MARK: - 测试结果 / Test Result

/// 测试结果结构体，记录单个测试的执行结果
/// EN: Test result struct recording the execution result of a single test
struct TestResult: Codable {
    /// 测试名称
    /// EN: Test name
    var test_name: String
    /// 编程语言
    /// EN: Programming language
    var language: String
    /// 测试模式
    /// EN: Test mode
    var mode: String
    /// 是否成功
    /// EN: Whether successful
    var success: Bool
    /// 错误信息（可选）
    /// EN: Error message (optional)
    var error: String?
    /// 执行耗时（毫秒）
    /// EN: Execution duration (milliseconds)
    var duration_ms: Int64
    /// 返回的文档列表（可选）
    /// EN: Returned document list (optional)
    var documents: [AnyCodable]?
    /// 文档数量（可选）
    /// EN: Document count (optional)
    var count: Int64?
    /// 匹配数量（可选）
    /// EN: Matched count (optional)
    var matched_count: Int64?
    /// 修改数量（可选）
    /// EN: Modified count (optional)
    var modified_count: Int64?
    /// 删除数量（可选）
    /// EN: Deleted count (optional)
    var deleted_count: Int64?
    /// 插入的文档ID（可选）
    /// EN: Upserted document ID (optional)
    var upserted_id: AnyCodable?
}

/// 结果文件结构体，包含所有测试结果和汇总信息
/// EN: Results file struct containing all test results and summary info
struct ResultsFile: Codable {
    /// 编程语言
    /// EN: Programming language
    let language: String
    /// 测试模式
    /// EN: Test mode
    let mode: String
    /// 测试结果列表
    /// EN: Test result list
    let results: [TestResult]
    /// 汇总信息
    /// EN: Summary info
    let summary: Summary
}

/// 汇总结构体，统计测试执行情况
/// EN: Summary struct for test execution statistics
struct Summary: Codable {
    /// 总测试数
    /// EN: Total test count
    let total: Int
    /// 通过数
    /// EN: Passed count
    let passed: Int
    /// 失败数
    /// EN: Failed count
    let failed: Int
    /// 跳过数
    /// EN: Skipped count
    let skipped: Int
}

// MARK: - 动态 JSON 处理 / Dynamic JSON Handling (AnyCodable)

/// AnyCodable 结构体，用于处理动态类型的 JSON 数据
/// EN: AnyCodable struct for handling dynamically typed JSON data
struct AnyCodable: Codable {
    /// 存储的实际值
    /// EN: The actual stored value
    let value: Any

    /// 使用任意值初始化
    /// EN: Initialize with any value
    init(_ value: Any) {
        self.value = value
    }

    /// 从解码器初始化，支持多种类型的自动解码
    /// EN: Initialize from decoder, supporting automatic decoding of multiple types
    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()

        if container.decodeNil() {
            value = NSNull()
        } else if let bool = try? container.decode(Bool.self) {
            value = bool
        } else if let int = try? container.decode(Int.self) {
            value = int
        } else if let int64 = try? container.decode(Int64.self) {
            value = int64
        } else if let double = try? container.decode(Double.self) {
            value = double
        } else if let string = try? container.decode(String.self) {
            value = string
        } else if let array = try? container.decode([AnyCodable].self) {
            value = array.map { $0.value }
        } else if let dict = try? container.decode([String: AnyCodable].self) {
            value = dict.mapValues { $0.value }
        } else {
            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Cannot decode AnyCodable")
        }
    }

    /// 编码到编码器，根据值类型自动选择编码方式
    /// EN: Encode to encoder, automatically selecting encoding method based on value type
    func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()

        switch value {
        case is NSNull:
            try container.encodeNil()
        case let bool as Bool:
            try container.encode(bool)
        case let int as Int:
            try container.encode(int)
        case let int64 as Int64:
            try container.encode(int64)
        case let double as Double:
            try container.encode(double)
        case let string as String:
            try container.encode(string)
        case let array as [Any]:
            try container.encode(array.map { AnyCodable($0) })
        case let dict as [String: Any]:
            try container.encode(dict.mapValues { AnyCodable($0) })
        default:
            let context = EncodingError.Context(codingPath: container.codingPath, debugDescription: "Cannot encode value")
            throw EncodingError.invalidValue(value, context)
        }
    }

    // MARK: - 类型转换辅助属性 / Type Conversion Helper Properties

    /// 转换为字典类型
    /// EN: Convert to dictionary type
    var dictionary: [String: Any]? {
        value as? [String: Any]
    }

    /// 转换为数组类型
    /// EN: Convert to array type
    var array: [Any]? {
        value as? [Any]
    }

    /// 转换为字符串类型
    /// EN: Convert to string type
    var string: String? {
        value as? String
    }

    /// 转换为整数类型
    /// EN: Convert to integer type
    var int: Int? {
        value as? Int
    }

    /// 转换为布尔类型
    /// EN: Convert to boolean type
    var bool: Bool? {
        value as? Bool
    }
}
