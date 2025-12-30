#!/bin/bash
# Created by Yanjunhui
# 运行 Go 测试

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

MODE=${1:-api}

echo "=== 运行 Go 测试 ($MODE 模式) ==="
cd "$PROJECT_DIR/runner/go"
go run . --mode=$MODE --output=../../reports/go_${MODE}.json
echo "完成"
