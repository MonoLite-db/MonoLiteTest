// Created by Yanjunhui

import Foundation

// MARK: - Test Case Definitions

struct TestCase: Codable {
    let name: String
    let category: String
    let operation: String
    let collection: String
    let description: String
    let setup: [SetupStep]?
    let action: TestAction
    let expected: Expected
}

struct SetupStep: Codable {
    let operation: String
    let data: AnyCodable
}

struct TestAction: Codable {
    let method: String
    let filter: AnyCodable?
    let update: AnyCodable?
    let doc: AnyCodable?
    let docs: [AnyCodable]?
    let options: AnyCodable?
}

struct Expected: Codable {
    let count: Int?
    let documents: [AnyCodable]?
    let matched_count: Int?
    let modified_count: Int?
    let deleted_count: Int?
    let upserted_id: AnyCodable?
    let error: String?
    let index_name: String?
}

struct TestSuite: Codable {
    let version: String
    let generated: String
    let tests: [TestCase]
}

// MARK: - Test Result

struct TestResult: Codable {
    var test_name: String
    var language: String
    var mode: String
    var success: Bool
    var error: String?
    var duration_ms: Int64
    var documents: [AnyCodable]?
    var count: Int64?
    var matched_count: Int64?
    var modified_count: Int64?
    var deleted_count: Int64?
    var upserted_id: AnyCodable?
}

struct ResultsFile: Codable {
    let language: String
    let mode: String
    let results: [TestResult]
    let summary: Summary
}

struct Summary: Codable {
    let total: Int
    let passed: Int
    let failed: Int
    let skipped: Int
}

// MARK: - AnyCodable (for dynamic JSON handling)

struct AnyCodable: Codable {
    let value: Any

    init(_ value: Any) {
        self.value = value
    }

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

    // Helper to convert to dictionary
    var dictionary: [String: Any]? {
        value as? [String: Any]
    }

    var array: [Any]? {
        value as? [Any]
    }

    var string: String? {
        value as? String
    }

    var int: Int? {
        value as? Int
    }

    var bool: Bool? {
        value as? Bool
    }
}
