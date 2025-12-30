#!/bin/bash
# Created by Yanjunhui
# 运行 Swift 测试

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

MODE=${1:-api}

echo "=== 运行 Swift 测试 ($MODE 模式) ==="
cd "$PROJECT_DIR/runner/swift"

# 构建（如果需要）
if [ ! -x ".build/debug/Runner" ]; then
    echo "构建 Swift 运行器..."
    swift build
fi

swift run Runner --mode=$MODE --output=../../reports/swift_${MODE}.json
echo "完成"
