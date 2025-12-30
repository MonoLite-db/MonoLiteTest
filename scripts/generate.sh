#!/bin/bash
# Created by Yanjunhui
# 生成测试数据

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== 生成测试数据 ==="
cd "$PROJECT_DIR/testdata/generator"
go run .
echo "完成"
