// Created by Yanjunhui

import Foundation
import MonoLiteSwift

#if canImport(Darwin)
import Darwin
#elseif canImport(Glibc)
import Glibc
#endif

actor WireRunner {
    private var db: Database?
    private var server: MongoWireTCPServer?
    private var clientFD: Int32 = -1
    private let dbPath: String
    private let port: Int
    private var requestId: Int32 = 1

    init(dbPath: String, port: Int) {
        self.dbPath = dbPath
        self.port = port
    }

    func open() async throws {
        // 打开数据库
        db = try await Database(path: dbPath)

        // 启动 Wire Protocol 服务器
        guard let db = db else { throw RunnerError.databaseNotOpened }
        server = try MongoWireTCPServer(addr: "127.0.0.1:\(port)", db: db)
        try await server?.start()

        // 等待服务器启动
        try await Task.sleep(nanoseconds: 200_000_000) // 200ms

        // 连接到服务器
        clientFD = socket(AF_INET, Int32(SOCK_STREAM), 0)
        guard clientFD >= 0 else {
            throw RunnerError.connectionFailed("Failed to create socket")
        }

        var addr = sockaddr_in()
        addr.sin_family = sa_family_t(AF_INET)
        addr.sin_port = UInt16(port).bigEndian
        addr.sin_addr = in_addr(s_addr: inet_addr("127.0.0.1"))

        var addrSock = sockaddr()
        memcpy(&addrSock, &addr, MemoryLayout<sockaddr_in>.size)

        let connectResult = withUnsafePointer(to: &addrSock) {
            $0.withMemoryRebound(to: sockaddr.self, capacity: 1) { ptr in
                connect(clientFD, ptr, socklen_t(MemoryLayout<sockaddr_in>.size))
            }
        }

        guard connectResult == 0 else {
            Darwin.close(clientFD)
            clientFD = -1
            throw RunnerError.connectionFailed("Failed to connect: \(errno)")
        }
    }

    func close() async throws {
        // 关闭客户端
        if clientFD >= 0 {
            Darwin.shutdown(clientFD, SHUT_RDWR)
            Darwin.close(clientFD)
            clientFD = -1
        }

        // 停止服务器
        await server?.stop()
        server = nil

        // 关闭数据库
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
            mode: "wire",
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

        for step in setup {
            let data = step.data.value as? [String: Any] ?? [:]

            switch step.operation {
            case "insert":
                var cmd = BSONDocument([
                    ("insert", .string(tc.collection)),
                    ("$db", .string("test"))
                ])
                let bsonDoc = convertToBSONDocument(data)
                cmd["documents"] = .array(BSONArray([.document(bsonDoc)]))
                _ = try await sendCommand(cmd)

            case "createIndex":
                if let keys = data["keys"] as? [String: Any] {
                    let keysDoc = convertToBSONDocument(keys)
                    var indexDoc = BSONDocument([("key", .document(keysDoc))])
                    if let opts = data["options"] as? [String: Any] {
                        if let name = opts["name"] as? String {
                            indexDoc["name"] = .string(name)
                        }
                    } else {
                        // 生成默认名称
                        var nameParts: [String] = []
                        for (k, v) in keysDoc {
                            nameParts.append("\(k)_\(v)")
                        }
                        indexDoc["name"] = .string(nameParts.joined(separator: "_"))
                    }

                    var cmd = BSONDocument([
                        ("createIndexes", .string(tc.collection)),
                        ("$db", .string("test"))
                    ])
                    cmd["indexes"] = .array(BSONArray([.document(indexDoc)]))
                    _ = try await sendCommand(cmd)
                }

            default:
                break
            }
        }
    }

    private func executeAction(_ tc: TestCase, result: inout TestResult) async throws {
        let action = tc.action
        let filterDict = action.filter?.value as? [String: Any] ?? [:]
        let updateDict = action.update?.value as? [String: Any] ?? [:]
        let docDict = action.doc?.value as? [String: Any] ?? [:]
        let options = action.options?.value as? [String: Any] ?? [:]

        let filter = convertToBSONDocument(filterDict)
        let update = convertToBSONDocument(updateDict)
        let doc = convertToBSONDocument(docDict)

        switch action.method {
        case "insertOne":
            var cmd = BSONDocument([
                ("insert", .string(tc.collection)),
                ("$db", .string("test"))
            ])
            cmd["documents"] = .array(BSONArray([.document(doc)]))
            _ = try await sendCommand(cmd)
            result.count = 1

        case "insertMany":
            if let docsArray = action.docs {
                var bsonDocs = BSONArray()
                for d in docsArray {
                    let dict = d.value as? [String: Any] ?? [:]
                    bsonDocs.append(.document(convertToBSONDocument(dict)))
                }
                var cmd = BSONDocument([
                    ("insert", .string(tc.collection)),
                    ("$db", .string("test"))
                ])
                cmd["documents"] = .array(bsonDocs)
                let resp = try await sendCommand(cmd)
                if let n = resp["n"]?.intValue {
                    result.count = Int64(n)
                }
            }

        case "find":
            var cmd = BSONDocument([
                ("find", .string(tc.collection)),
                ("$db", .string("test")),
                ("filter", .document(filter))
            ])
            if let sortVal = options["sort"] as? [String: Any] {
                cmd["sort"] = .document(convertToBSONDocument(sortVal))
            }
            if let limit = options["limit"] as? Int {
                cmd["limit"] = .int32(Int32(limit))
            }
            if let skip = options["skip"] as? Int {
                cmd["skip"] = .int32(Int32(skip))
            }
            if let projVal = options["projection"] as? [String: Any] {
                cmd["projection"] = .document(convertToBSONDocument(projVal))
            }

            let resp = try await sendCommand(cmd)
            if let cursor = resp["cursor"]?.documentValue,
               let firstBatch = cursor["firstBatch"]?.arrayValue {
                result.count = Int64(firstBatch.count)
            }

        case "findOne":
            var cmd = BSONDocument([
                ("find", .string(tc.collection)),
                ("$db", .string("test")),
                ("filter", .document(filter)),
                ("limit", .int32(1))
            ])
            let resp = try await sendCommand(cmd)
            if let cursor = resp["cursor"]?.documentValue,
               let firstBatch = cursor["firstBatch"]?.arrayValue {
                result.count = firstBatch.isEmpty ? 0 : 1
            }

        case "updateOne":
            let upsert = options["upsert"] as? Bool ?? false
            var updateDoc = BSONDocument([
                ("q", .document(filter)),
                ("u", .document(update)),
                ("upsert", .bool(upsert))
            ])
            var cmd = BSONDocument([
                ("update", .string(tc.collection)),
                ("$db", .string("test"))
            ])
            cmd["updates"] = .array(BSONArray([.document(updateDoc)]))
            let resp = try await sendCommand(cmd)
            result.matched_count = Int64(resp["n"]?.intValue ?? 0)
            result.modified_count = Int64(resp["nModified"]?.intValue ?? 0)

        case "updateMany":
            var updateDoc = BSONDocument([
                ("q", .document(filter)),
                ("u", .document(update)),
                ("multi", .bool(true))
            ])
            var cmd = BSONDocument([
                ("update", .string(tc.collection)),
                ("$db", .string("test"))
            ])
            cmd["updates"] = .array(BSONArray([.document(updateDoc)]))
            let resp = try await sendCommand(cmd)
            result.matched_count = Int64(resp["n"]?.intValue ?? 0)
            result.modified_count = Int64(resp["nModified"]?.intValue ?? 0)

        case "deleteOne":
            var deleteDoc = BSONDocument([
                ("q", .document(filter)),
                ("limit", .int32(1))
            ])
            var cmd = BSONDocument([
                ("delete", .string(tc.collection)),
                ("$db", .string("test"))
            ])
            cmd["deletes"] = .array(BSONArray([.document(deleteDoc)]))
            let resp = try await sendCommand(cmd)
            result.deleted_count = Int64(resp["n"]?.intValue ?? 0)

        case "deleteMany":
            var deleteDoc = BSONDocument([
                ("q", .document(filter)),
                ("limit", .int32(0))
            ])
            var cmd = BSONDocument([
                ("delete", .string(tc.collection)),
                ("$db", .string("test"))
            ])
            cmd["deletes"] = .array(BSONArray([.document(deleteDoc)]))
            let resp = try await sendCommand(cmd)
            result.deleted_count = Int64(resp["n"]?.intValue ?? 0)

        case "replaceOne":
            // replaceOne 使用 update 命令，但 u 不带 $ 操作符
            var updateDoc = BSONDocument([
                ("q", .document(filter)),
                ("u", .document(doc))
            ])
            var cmd = BSONDocument([
                ("update", .string(tc.collection)),
                ("$db", .string("test"))
            ])
            cmd["updates"] = .array(BSONArray([.document(updateDoc)]))
            let resp = try await sendCommand(cmd)
            result.matched_count = Int64(resp["n"]?.intValue ?? 0)
            result.modified_count = Int64(resp["nModified"]?.intValue ?? 0)

        case "distinct":
            let field = options["field"] as? String ?? ""
            var cmd = BSONDocument([
                ("distinct", .string(tc.collection)),
                ("$db", .string("test")),
                ("key", .string(field)),
                ("query", .document(filter))
            ])
            let resp = try await sendCommand(cmd)
            if let values = resp["values"]?.arrayValue {
                result.count = Int64(values.count)
            }

        case "aggregate":
            if let pipeline = options["pipeline"] as? [[String: Any]] {
                var stages = BSONArray()
                for stage in pipeline {
                    stages.append(.document(convertToBSONDocument(stage)))
                }
                var cmd = BSONDocument([
                    ("aggregate", .string(tc.collection)),
                    ("$db", .string("test")),
                    ("pipeline", .array(stages)),
                    ("cursor", .document(BSONDocument()))
                ])
                let resp = try await sendCommand(cmd)
                if let cursor = resp["cursor"]?.documentValue,
                   let firstBatch = cursor["firstBatch"]?.arrayValue {
                    result.count = Int64(firstBatch.count)
                }
            }

        case "createIndex":
            if let keys = options["keys"] as? [String: Any] {
                let keysDoc = convertToBSONDocument(keys)
                var indexDoc = BSONDocument([("key", .document(keysDoc))])
                if let indexOpts = options["options"] as? [String: Any] {
                    if let name = indexOpts["name"] as? String {
                        indexDoc["name"] = .string(name)
                    }
                    if let unique = indexOpts["unique"] as? Bool {
                        indexDoc["unique"] = .bool(unique)
                    }
                }
                // 生成默认名称如果没有
                if indexDoc["name"] == nil {
                    var nameParts: [String] = []
                    for (k, v) in keysDoc {
                        nameParts.append("\(k)_\(v)")
                    }
                    indexDoc["name"] = .string(nameParts.joined(separator: "_"))
                }

                var cmd = BSONDocument([
                    ("createIndexes", .string(tc.collection)),
                    ("$db", .string("test"))
                ])
                cmd["indexes"] = .array(BSONArray([.document(indexDoc)]))
                _ = try await sendCommand(cmd)
                result.count = 1
            }

        case "listIndexes":
            var cmd = BSONDocument([
                ("listIndexes", .string(tc.collection)),
                ("$db", .string("test"))
            ])
            let resp = try await sendCommand(cmd)
            if let cursor = resp["cursor"]?.documentValue,
               let firstBatch = cursor["firstBatch"]?.arrayValue {
                result.count = Int64(firstBatch.count)
            }

        case "dropIndex":
            if let name = options["name"] as? String {
                var cmd = BSONDocument([
                    ("dropIndexes", .string(tc.collection)),
                    ("$db", .string("test")),
                    ("index", .string(name))
                ])
                _ = try await sendCommand(cmd)
            }

        default:
            throw RunnerError.unknownMethod(action.method)
        }
    }

    // MARK: - Wire Protocol Communication

    private func sendCommand(_ cmd: BSONDocument) async throws -> BSONDocument {
        guard clientFD >= 0 else { throw RunnerError.connectionFailed("Not connected") }

        // 构建 OP_MSG 消息
        let encoder = BSONEncoder()
        let bsonData = try encoder.encode(cmd)

        // OP_MSG 格式: flagBits(4) + kind(1) + document
        var opMsgBody = Data()
        opMsgBody.append(contentsOf: [0, 0, 0, 0]) // flagBits = 0
        opMsgBody.append(0) // kind = 0 (body)
        opMsgBody.append(bsonData)

        // Wire Message Header: length(4) + requestId(4) + responseTo(4) + opCode(4)
        let msgLen = Int32(16 + opMsgBody.count)
        let reqId = requestId
        requestId += 1

        var header = Data(count: 16)
        writeInt32LE(&header, at: 0, value: msgLen)
        writeInt32LE(&header, at: 4, value: reqId)
        writeInt32LE(&header, at: 8, value: 0) // responseTo
        writeInt32LE(&header, at: 12, value: 2013) // OP_MSG

        var fullMsg = Data()
        fullMsg.append(header)
        fullMsg.append(opMsgBody)

        // 发送消息
        try sendAll(fullMsg)

        // 接收响应
        let respHeader = try recvExact(16)
        let respLen = Int(readInt32LE(respHeader, at: 0))
        guard respLen >= 16, respLen < 48 * 1024 * 1024 else {
            throw RunnerError.connectionFailed("Invalid response length: \(respLen)")
        }

        let respBody = try recvExact(respLen - 16)

        // 解析 OP_MSG 响应
        // flagBits(4) + kind(1) + document
        guard respBody.count >= 5 else {
            throw RunnerError.connectionFailed("Response too short")
        }

        let docStart = 5 // skip flagBits + kind
        let docData = Data(respBody[docStart...])

        let decoder = BSONDecoder()
        return try decoder.decode(docData)
    }

    private func sendAll(_ data: Data) throws {
        var sent = 0
        while sent < data.count {
            let n = data.withUnsafeBytes { ptr in
                let base = ptr.baseAddress!.assumingMemoryBound(to: UInt8.self).advanced(by: sent)
                return send(clientFD, base, data.count - sent, 0)
            }
            if n <= 0 {
                throw RunnerError.connectionFailed("send failed: \(errno)")
            }
            sent += n
        }
    }

    private func recvExact(_ count: Int) throws -> Data {
        var buf = Data(count: count)
        var received = 0
        while received < count {
            let n = buf.withUnsafeMutableBytes { ptr in
                let base = ptr.baseAddress!.assumingMemoryBound(to: UInt8.self).advanced(by: received)
                return recv(clientFD, base, count - received, 0)
            }
            if n == 0 {
                throw RunnerError.connectionFailed("connection closed")
            }
            if n < 0 {
                throw RunnerError.connectionFailed("recv failed: \(errno)")
            }
            received += n
        }
        return buf
    }

    // MARK: - Helpers

    private func writeInt32LE(_ data: inout Data, at offset: Int, value: Int32) {
        var v = value.littleEndian
        withUnsafeBytes(of: &v) { bytes in
            for i in 0..<4 {
                data[offset + i] = bytes[i]
            }
        }
    }

    private func readInt32LE(_ data: Data, at offset: Int) -> Int32 {
        var result: Int32 = 0
        withUnsafeMutableBytes(of: &result) { bytes in
            for i in 0..<4 {
                bytes[i] = data[offset + i]
            }
        }
        return Int32(littleEndian: result)
    }

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
