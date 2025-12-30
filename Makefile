# Created by Yanjunhui
# MonoLite 三语言一致性测试

.PHONY: all generate test-go test-swift test-ts verify report clean

# 默认目标：运行完整测试流程
all: generate test-go test-swift test-ts verify

# 生成测试数据
generate:
	@echo "=== 生成测试数据 ==="
	cd testdata/generator && go run .

# Go 测试
test-go:
	@echo "=== 运行 Go 测试 ==="
	cd runner/go && go run . --mode=api --output=../../reports/go_api.json
	cd runner/go && go run . --mode=wire --output=../../reports/go_wire.json

# Swift 测试
test-swift:
	@echo "=== 运行 Swift 测试 ==="
	cd runner/swift && swift run Runner --mode api --output ../../reports/swift_api.json
	cd runner/swift && swift run Runner --mode wire --output ../../reports/swift_wire.json

# TypeScript 测试
test-ts:
	@echo "=== 运行 TypeScript 测试 ==="
	cd runner/ts && npm run start -- --mode=api --output=../../reports/ts_api.json
	cd runner/ts && npm run start -- --mode=wire --output=../../reports/ts_wire.json

# 验证和报告生成
verify:
	@echo "=== 生成一致性报告 ==="
	cd verifier && go run .

# 清理
clean:
	rm -f testdata/fixtures/test.monodb
	rm -f testdata/fixtures/testcases.json
	rm -rf testdata/fixtures/expected/*
	rm -f reports/*.json
	rm -f reports/*.md

# 仅测试 API 模式
test-api: generate
	cd runner/go && go run . --mode=api --output=../../reports/go_api.json
	cd runner/swift && swift run Runner --mode api --output ../../reports/swift_api.json
	cd runner/ts && npm run start -- --mode=api --output=../../reports/ts_api.json

# 仅测试 Wire 模式
test-wire: generate
	cd runner/go && go run . --mode=wire --output=../../reports/go_wire.json
	cd runner/swift && swift run Runner --mode wire --output ../../reports/swift_wire.json
	cd runner/ts && npm run start -- --mode=wire --output=../../reports/ts_wire.json

# 安装依赖
deps:
	cd testdata/generator && go mod tidy
	cd runner/go && go mod tidy
	cd runner/swift && swift package resolve
	cd runner/ts && npm install
	cd verifier && go mod tidy

# 快速测试（仅 Go API 模式）
quick:
	@echo "=== 快速测试 ==="
	cd testdata/generator && go run .
	cd runner/go && go run . --mode=api --output=../../reports/go_api.json

# 帮助
help:
	@echo "MonoLite 三语言一致性测试"
	@echo ""
	@echo "命令:"
	@echo "  make all        - 运行完整测试流程"
	@echo "  make generate   - 生成测试数据"
	@echo "  make test-go    - 运行 Go 测试"
	@echo "  make test-swift - 运行 Swift 测试"
	@echo "  make test-ts    - 运行 TypeScript 测试"
	@echo "  make verify     - 生成一致性报告"
	@echo "  make clean      - 清理生成的文件"
	@echo "  make deps       - 安装依赖"
	@echo "  make quick      - 快速测试（仅 Go API）"
