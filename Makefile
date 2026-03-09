APP_NAME := q-deploy
BUILD_DIR := backend
BINARY := $(BUILD_DIR)/bin/$(APP_NAME)

# 热加载调试的子命令，默认 server，可通过 make dev CMD=xxx 覆盖
CMD ?= server

.PHONY: init build run dev swagger sql lint test cover fe-install fe-dev fe-build fe-lint infra-up infra-down infra-logs migrate docker-build docker-up docker-down clean

# ---------- 初始化 ----------

init:
	@echo "📦 安装后端工具..."
	cd $(BUILD_DIR) && go install github.com/air-verse/air@latest
	cd $(BUILD_DIR) && go install github.com/swaggo/swag/cmd/swag@latest
	cd $(BUILD_DIR) && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "📦 安装前端依赖..."
	cd frontend && pnpm install
	@echo ""
	@echo "✅ 初始化完成！"
	@echo ""
	@echo "接下来:"
	@echo "  make infra-up   # 启动基础设施"
	@echo "  make dev       # 启动后端 (http://localhost:8080)"
	@echo "  make fe-dev    # 启动前端 (http://localhost:8000)"
	@echo ""

# ---------- 构建 & 运行 ----------

build:
	cd $(BUILD_DIR) && go build -o bin/$(APP_NAME) .

run: build
	cd $(BUILD_DIR) && ./bin/$(APP_NAME) server

# ---------- 热加载调试 ----------

dev:
	cd $(BUILD_DIR) && air -- $(CMD)

# ---------- 代码生成 ----------

swagger:
	cd $(BUILD_DIR) && go run github.com/swaggo/swag/cmd/swag init -o ./gen/docs --parseDependency

sql:
	cd $(BUILD_DIR) && go run ./gen/gorm_gen

# ---------- 代码检查 ----------

lint:
	cd $(BUILD_DIR) && go vet ./...
	cd $(BUILD_DIR) && golangci-lint run ./...

# ---------- 测试 ----------

test:
	cd $(BUILD_DIR) && go test ./... -v -count=1

cover:
	cd $(BUILD_DIR) && go test ./... -coverprofile=coverage.out -count=1
	cd $(BUILD_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: $(BUILD_DIR)/coverage.html"

# ---------- 前端 ----------

fe-install:
	cd frontend && pnpm install

fe-dev:
	cd frontend && pnpm dev

fe-build:
	cd frontend && pnpm build

fe-lint:
	cd frontend && pnpm lint
	cd frontend && pnpm format:check

# ---------- 基础设施（本地调试） ----------

COMPOSE := docker compose -f deploy/docker-compose.yml

migrate:
	@echo "🔄 执行数据库迁移..."
	cd $(BUILD_DIR) && go run . migrate

infra-up:
	$(COMPOSE) up -d
	@echo ""
	@echo "⏳ 等待 MySQL 就绪..."
	@for i in $$(seq 30); do \
		docker exec $$(docker ps -qf "name=mysql" 2>/dev/null) mysqladmin ping -h localhost -uroot -proot 2>/dev/null && break; \
		sleep 1; \
	done
	@echo "🔄 执行数据库迁移..."
	cd $(BUILD_DIR) && go run . migrate
	@echo ""
	@echo "✅ 基础设施启动完成！"
	@echo ""
	@echo "📌 服务地址:"
	@echo "   MySQL:      localhost:3306"
	@echo "   Redis:      localhost:6379"
	@echo "   Kafka:      localhost:9092"
	@echo "   Jaeger:     http://localhost:16686"
	@echo "   Prometheus: http://localhost:9090"
	@echo ""
	@echo "🚀 本地开发请另开终端运行:"
	@echo "   make dev      # 后端 (http://localhost:8080)"
	@echo "   make fe-dev   # 前端 (http://localhost:8000)"
	@echo ""

infra-down:
	$(COMPOSE) down

infra-logs:
	$(COMPOSE) logs -f

# ---------- Docker（全栈部署） ----------

docker-build:
	docker build -t $(APP_NAME) -f deploy/backend-Dockerfile .
	docker build -t $(APP_NAME)-web -f deploy/frontend-Dockerfile .

docker-up:
	$(COMPOSE) --profile deploy up -d
	@echo ""
	@echo "✅ 全栈服务启动完成！"
	@echo ""
	@echo "📌 访问地址:"
	@echo "   前端:       http://localhost"
	@echo "   后端 API:   http://localhost:8080"
	@echo "   Jaeger:     http://localhost:16686"
	@echo "   Prometheus: http://localhost:9090"
	@echo ""
	@echo "🔍 健康检查:"
	@echo "   curl http://localhost:8080/healthz"
	@echo "   curl http://localhost:8080/readyz"
	@echo ""

docker-down:
	$(COMPOSE) --profile deploy down

# ---------- 清理 ----------

clean:
	rm -rf $(BUILD_DIR)/bin
