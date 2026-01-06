package handlers

import (
	"fmt"
	"nosql_db/internal/api"
	"nosql_db/internal/storage"
)

func handleCreateIndex(req api.Request) api.Response {
	fieldName := ""
	for k := range req.Query {
		fieldName = k
		break
	}

	if fieldName == "" {
		return api.Response{Status: api.StatusError, Message: "field name required in query"}
	}

	// Используем очередь для write-операции
	result := storage.GlobalManager.Enqueue(req.Database, func(coll *storage.Collection) (storage.WriteResult, error) {
		if err := coll.CreateIndex(fieldName, 64); err != nil {
			return storage.WriteResult{}, fmt.Errorf("failed to create index: %w", err)
		}

		return storage.WriteResult{
			Message: fmt.Sprintf("Index created on field '%s'", fieldName),
		}, nil
	})

	if result.Error != nil {
		return api.Response{Status: api.StatusError, Message: result.Error.Error()}
	}

	return api.Response{
		Status:  api.StatusSuccess,
		Message: result.Message,
	}
}
