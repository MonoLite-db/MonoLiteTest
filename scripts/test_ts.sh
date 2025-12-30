#!/bin/bash
# Created by Yanjunhui
# 运行 TypeScript 测试

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

MODE=${1:-api}

echo "=== 运行 TypeScript 测试 ($MODE 模式) ==="
cd "$PROJECT_DIR/runner/ts"

# 安装依赖（如果需要）
if [ ! -d "node_modules" ]; then
    echo "安装依赖..."
    npm install
fi

npx ts-node src/index.ts --mode=$MODE --output=../../reports/ts_${MODE}.json
echo "完成"
