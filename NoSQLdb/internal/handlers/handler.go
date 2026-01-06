package handlers

import (
	"fmt"
	"nosql_db/internal/api"
	"nosql_db/internal/storage"
)

// HandleRequest — точка входа для обработки запросов
func HandleRequest(req api.Request) api.Response {
	if req.Database == "" {
		return api.Response{Status: api.StatusError, Message: "database name is required"}
	}

	switch req.Command {
	case api.CmdInsert:
		// Write-операция через очередь
		return handleInsert(req)
	case api.CmdFind:
		// Read-операция напрямую (не требует очереди)
		coll, err := storage.GlobalManager.GetCollection(req.Database)
		if err != nil {
			return api.Response{Status: api.StatusError, Message: fmt.Sprintf("failed to load database: %v", err)}
		}
		return handleFind(coll, req)
	case api.CmdDelete:
		// Write-операция через очередь
		return handleDelete(req)
	case api.CmdCreateIndex:
		// Write-операция через очередь
		return handleCreateIndex(req)
	default:
		return api.Response{Status: api.StatusError, Message: fmt.Sprintf("unknown command: %s", req.Command)}
	}
}
