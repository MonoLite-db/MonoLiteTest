// Created by Yanjunhui

import Foundation
import MonoLiteSwift

actor APIRunner {
    private var db: Database?
    private let dbPath: String

    init(dbPath: String) {
        self.dbPath = dbPath
    }

    func open() async throws {
        db = try await Database(path: dbPath)
    }

    func close() async throws {
        if let db = db {
            try await db.close()
            self.db = nil
        }
    }

    func runTest(_ tc: TestCase) async -> TestResult {
        let start = Date()
        var result = TestResult(
            test_name: tc.name,
            language: "swift",
            mode: "api",
            success: false,
            duration_ms: 0
        )

        do {
            // 执行前置步骤
            try await executeSetup(tc)

            // 执行测试动作
            try await executeAction(tc, result: &result)

            result.success = true
        } catch {
            result.error = "\(error)"

            // 检查是否是预期的错误
            if let expectedError = tc.expected.error, result.error?.contains(expectedError) == true {
                result.success = true
            }
        }

        result.duration_ms = Int64(Date().timeIntervalSince(start) * 1000)
        return result
    }

    private func executeSetup(_ tc: TestCase) async throws {
        guard let setup = tc.setup, !setup.isEmpty else { return }
        guard let db = db else { throw RunnerError.databaseNotOpened }

        let col = try await db.collection(tc.collection)

        for step in setup {
            let data = convertToBSONDocument(step.data.value)

            switch step.operation {
            case "insert":
                _ = try await col.insertOne(data)
            case "createIndex":
                if let keys = (step.data.value as? [String: Any])?["keys"] as? [String: Any] {
                    let keysDoc = convertToBSONDocument(keys)
                    let optsDoc = BSONDocument()
                    _ = try await col.createIndex(keys: keysDoc, options: optsDoc)
                }
            default:
                break
            }
        }
    }

    private func executeAction(_ tc: TestCase, result: inout TestResult) async throws {
        guard let db = db else { throw RunnerError.databaseNotOpened }

        let col = try await db.collection(tc.collection)
        let action = tc.action

        let filter = action.filter.map { convertToBSONDocument($0.value) } ?? BSONDocument()
        let update = action.update.map { convertToBSONDocument($0.value) } ?? BSONDocument()
        let doc = action.doc.map { convertToBSONDocument($0.value) } ?? BSONDocument()
        let options = action.options?.value as? [String: Any] ?? [:]

        switch action.method {
        case "insertOne":
            _ = try await col.insertOne(doc)
            result.count = 1

        case "insertMany":
            if let docsArray = action.docs {
                let bsonDocs = docsArray.map { convertToBSONDocument($0.value) }
                let ids = try await col.insert(bsonDocs)
                result.count = Int64(ids.count)
            }

        case "find":
            var queryOpts = QueryOptions()
            if let sortVal = options["sort"] as? [String: Any] {
                queryOpts.sort = convertToBSONDocument(sortVal)
            }
            if let limit = options["limit"] as? Int {
                queryOpts.limit = Int64(limit)
            }
            if let skip = options["skip"] as? Int {
                queryOpts.skip = Int64(skip)
            }
            if let projVal = options["projection"] as? [String: Any] {
                queryOpts.projection = convertToBSONDocument(projVal)
            }

            let docs: [BSONDocument]
            if !queryOpts.sort.isEmpty || queryOpts.limit > 0 || queryOpts.skip > 0 || !queryOpts.projection.isEmpty {
                docs = try await col.findWithOptions(filter, options: queryOpts)
            } else {
                docs = try await col.find(filter)
            }
            result.count = Int64(docs.count)

        case "findOne":
            let doc = try await col.findOne(filter)
            result.count = doc != nil ? 1 : 0

        case "updateOne":
            let upsert = options["upsert"] as? Bool ?? false
            let res = try await col.updateOne(filter, update: update, upsert: upsert)
            result.matched_count = res.matchedCount
            result.modified_count = res.modifiedCount

        case "updateMany":
            let res = try await col.update(filter, update: update, upsert: false)
            result.matched_count = res.matchedCount
            result.modified_count = res.modifiedCount

        case "deleteOne":
            let count = try await col.deleteOne(filter)
            result.deleted_count = count

        case "deleteMany":
            let count = try await col.delete(filter)
            result.deleted_count = count

        case "replaceOne":
            let count = try await col.replaceOne(filter, replacement: doc)
            result.matched_count = count > 0 ? 1 : 0
            result.modified_count = count

        case "findAndModify":
            var cmd = BSONDocument()
            cmd["findAndModify"] = .string(tc.collection)
            cmd["query"] = .document(filter)
            if !update.isEmpty {
                cmd["update"] = .document(update)
            }
            if let newVal = options["new"] as? Bool {
                cmd["new"] = .bool(newVal)
            }
            if let upsert = options["upsert"] as? Bool {
                cmd["upsert"] = .bool(upsert)
            }
            if let remove = options["remove"] as? Bool {
                cmd["remove"] = .bool(remove)
            }
            let res = try await db.runCommand(cmd)
            if res["value"] != nil && res["value"] != .null {
                result.count = 1
            } else {
                result.count = 0
            }

        case "distinct":
            let field = options["field"] as? String ?? ""
            let values = try await col.distinct(field, filter: filter)
            result.count = Int64(values.count)

        case "aggregate":
            if let pipeline = options["pipeline"] as? [[String: Any]] {
                let stages = pipeline.map { convertToBSONDocument($0) }
                let pipelineObj = try Pipeline(stages: stages, db: db)
                let inputDocs = try await col.find(BSONDocument())
                let outDocs = try await pipelineObj.execute(inputDocs)
                result.count = Int64(outDocs.count)
            }

        case "createIndex":
            if let keys = options["keys"] as? [String: Any] {
                let keysDoc = convertToBSONDocument(keys)
                var optsDoc = BSONDocument()
                if let indexOpts = options["options"] as? [String: Any] {
                    if let name = indexOpts["name"] as? String {
                        optsDoc["name"] = .string(name)
                    }
                    if let unique = indexOpts["unique"] as? Bool {
                        optsDoc["unique"] = .bool(unique)
                    }
                }
                _ = try await col.createIndex(keys: keysDoc, options: optsDoc)
                result.count = 1
            }

        case "listIndexes":
            let indexes = await col.listIndexes()
            result.count = Int64(indexes.count)

        case "dropIndex":
            if let name = options["name"] as? String {
                try await col.dropIndex(name)
            }

        default:
            throw RunnerError.unknownMethod(action.method)
        }
    }

    // MARK: - BSON Conversion

    private func convertToBSONDocument(_ value: Any) -> BSONDocument {
        guard let dict = value as? [String: Any] else {
            return BSONDocument()
        }

        var doc = BSONDocument()
        for (key, val) in dict {
            doc[key] = convertToBSONValue(val)
        }
        return doc
    }

    private func convertToBSONValue(_ value: Any) -> BSONValue {
        switch value {
        case is NSNull:
            return .null
        case let bool as Bool:
            return .bool(bool)
        case let int as Int:
            return .int64(Int64(int))
        case let int64 as Int64:
            return .int64(int64)
        case let double as Double:
            return .double(double)
        case let string as String:
            return .string(string)
        case let array as [Any]:
            let bsonArray = BSONArray(array.map { convertToBSONValue($0) })
            return .array(bsonArray)
        case let dict as [String: Any]:
            return .document(convertToBSONDocument(dict))
        default:
            return .null
        }
    }
}

enum RunnerError: Error, CustomStringConvertible {
    case databaseNotOpened
    case unknownMethod(String)
    case connectionFailed(String)

    var description: String {
        switch self {
        case .databaseNotOpened:
            return "Database not opened"
        case .unknownMethod(let method):
            return "Unknown method: \(method)"
        case .connectionFailed(let msg):
            return "Connection failed: \(msg)"
        }
    }
}
