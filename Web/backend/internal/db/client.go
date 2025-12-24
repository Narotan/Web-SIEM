package db

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/Narotan/Web-SIEM/internal/config"
)

func SendQuery(req DBRequest) (*DBResponse, error) {
	cfg := config.GetConfig()

	conn, err := net.DialTimeout("tcp", cfg.DBAddr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к СУБД по адресу %s: %w", cfg.DBAddr, err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(req); err != nil {
		return nil, fmt.Errorf("ошибка кодирования запроса: %w", err)
	}

	var resp DBResponse
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&resp); err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа от СУБД: %w", err)
	}

	return &resp, nil
}
