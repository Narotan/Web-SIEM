package handlers

import (
	"fmt"
	"nosql_db/internal/api"
	"nosql_db/internal/storage"
)

func handleInsert(req api.Request) api.Response {
	if len(req.Data) == 0 {
		return api.Response{Status: api.StatusError, Message: "no data provided for insert"}
	}

	// Используем очередь для write-операции
	result := storage.GlobalManager.Enqueue(req.Database, func(coll *storage.Collection) (storage.WriteResult, error) {
		var insertedIDs []string

		for _, doc := range req.Data {
			id, err := coll.Insert(doc)
			if err != nil {
				return storage.WriteResult{}, fmt.Errorf("insert error: %w", err)
			}
			insertedIDs = append(insertedIDs, id)
		}

		if err := coll.Save(); err != nil {
			return storage.WriteResult{}, fmt.Errorf("failed to save data: %w", err)
		}

		if err := coll.SaveAllIndexes(); err != nil {
			return storage.WriteResult{}, fmt.Errorf("failed to save indexes: %w", err)
		}

		return storage.WriteResult{
			InsertedIDs: insertedIDs,
			Message:     fmt.Sprintf("Inserted %d document(s)", len(insertedIDs)),
		}, nil
	})

	if result.Error != nil {
		return api.Response{Status: api.StatusError, Message: result.Error.Error()}
	}

	return api.Response{
		Status:  api.StatusSuccess,
		Message: result.Message,
		Count:   len(result.InsertedIDs),
	}
}
