// swift-tools-version: 5.9
// Created by Yanjunhui

import PackageDescription

let package = Package(
    name: "SwiftTestRunner",
    platforms: [
        .macOS(.v13)
    ],
    dependencies: [
        .package(path: "../../../MonoLiteSwift"),
        .package(url: "https://github.com/apple/swift-argument-parser.git", from: "1.3.0")
    ],
    targets: [
        .executableTarget(
            name: "Runner",
            dependencies: [
                .product(name: "MonoLiteSwift", package: "MonoLiteSwift"),
                .product(name: "ArgumentParser", package: "swift-argument-parser")
            ],
            path: "Sources/Runner"
        )
    ]
)
