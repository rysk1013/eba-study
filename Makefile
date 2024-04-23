# makeを打った時のコマンド
.DEFAULT_GOAL := help

.PHONY: build
build: ## ビルド
	@docker compose -f docker-compose-prod.yml build --no-cache

.PHONY: up
up: ## コンテナ起動
	@docker compose -f docker-compose-prod.yml up -d

.PHONY: down
down: ## コンテナ停止・削除
	@docker compose -f docker-compose-prod.yml down

.PHONY: help
help: ## ヘルプ
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'