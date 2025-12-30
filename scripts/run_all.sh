#!/bin/bash
# Created by Yanjunhui
# MonoLite 三语言一致性测试完整流程

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== MonoLite 三语言一致性测试 ==="
echo "时间: $(date)"
echo "项目目录: $PROJECT_DIR"
echo ""

# Step 1: 生成测试数据
echo "[1/5] 生成测试数据..."
cd "$PROJECT_DIR/testdata/generator"
go run .
echo ""

# Step 2: 运行 Go 测试
echo "[2/5] 运行 Go 测试..."
cd "$PROJECT_DIR/runner/go"

echo "  API 模式..."
go run . --mode=api --output=../../reports/go_api.json

echo "  Wire 模式..."
go run . --mode=wire --output=../../reports/go_wire.json
echo ""

# Step 3: 运行 TypeScript 测试
echo "[3/5] 运行 TypeScript 测试..."
cd "$PROJECT_DIR/runner/ts"

# 安装依赖（如果需要）
if [ ! -d "node_modules" ]; then
    echo "  安装依赖..."
    npm install
fi

echo "  API 模式..."
npx ts-node src/index.ts --mode=api --output=../../reports/ts_api.json

echo "  Wire 模式..."
npx ts-node src/index.ts --mode=wire --output=../../reports/ts_wire.json
echo ""

# Step 4: 运行 Swift 测试（如果可用）
echo "[4/5] 运行 Swift 测试..."
cd "$PROJECT_DIR/runner/swift"
if [ -f "Package.swift" ]; then
    echo "  构建 Swift 运行器..."
    swift build 2>/dev/null || echo "  警告: Swift 构建失败，跳过"

    if [ -x ".build/debug/Runner" ]; then
        echo "  API 模式..."
        swift run Runner --mode api --output ../../reports/swift_api.json 2>/dev/null || echo "  警告: Swift API 测试失败"

        echo "  Wire 模式..."
        swift run Runner --mode wire --output ../../reports/swift_wire.json 2>/dev/null || echo "  警告: Swift Wire 测试失败"
    else
        echo "  跳过 Swift 测试（未构建成功）"
    fi
else
    echo "  跳过 Swift 测试（Package.swift 不存在）"
fi
echo ""

# Step 5: 生成报告
echo "[5/5] 生成一致性报告..."
cd "$PROJECT_DIR/verifier"
go run .
echo ""

echo "=== 测试完成 ==="
echo "报告位置:"
echo "  JSON: $PROJECT_DIR/reports/consistency_report.json"
echo "  Markdown: $PROJECT_DIR/reports/consistency_report.md"
