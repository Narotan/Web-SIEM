# Переменные путей
DB_DIR=NoSQLdb
AGENT_DIR=SIEM-Agent
WEB_DIR=Web/backend
BIN_DIR=bin

# Цвета для вывода
BLUE=\033[0;34m
GREEN=\033[0;32m
NC=\033[0m # No Color

.PHONY: all build run-db run-agent run-web clean help

help:
	@echo -e "$(BLUE)Доступные команды:$(NC)"
	@echo "  make build      - Собрать все модули в папку bin/"
	@echo "  make run-db     - Запустить NoSQLdb"
	@echo "  make run-web    - Запустить Web-Backend"
	@echo "  make run-agent  - Запустить SIEM-Agent (требуется sudo)"
	@echo "  make clean      - Очистить бинарники и временные логи"

# Сборка всех компонентов
build:
	@echo -e "$(GREEN)Сборка всех компонентов...$(NC)"
	@mkdir -p $(BIN_DIR)
	cd $(DB_DIR) && go build -o ../$(BIN_DIR)/nosql_db ./cmd/server/main.go
	cd $(AGENT_DIR) && go build -o ../$(BIN_DIR)/siem-agent ./cmd/agent/main.go
	cd $(WEB_DIR) && go build -o ../$(BIN_DIR)/web-siem ./cmd/server/main.go

# Запуск БД (Должна быть запущена первой)
run-db:
	@echo -e "$(GREEN)Запуск NoSQLdb на порту 5140...$(NC)"
	cd $(DB_DIR) && go run ./cmd/server/main.go

# Запуск Веб-сервера
run-web:
	@echo -e "$(GREEN)Запуск Web-Backend на порту 8000...$(NC)"
	cd $(WEB_DIR) && go run ./cmd/server/main.go

# Запуск Агента (нужен sudo для доступа к /var/log/audit)
run-agent:
	@echo -e "$(GREEN)Запуск SIEM-Agent...$(NC)"
	cd $(AGENT_DIR) && sudo go run ./cmd/agent/main.go

# Очистка системы
clean:
	@echo -e "$(BLUE)Очистка...$(NC)"
	rm -rf $(BIN_DIR)
	rm -rf $(AGENT_DIR)/storage/buffer/*
	rm -f $(WEB_DIR)/web-siem
	@echo "Готово."