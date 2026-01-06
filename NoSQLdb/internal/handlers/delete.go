package handlers

import (
	"fmt"
	"nosql_db/internal/api"
	"nosql_db/internal/operators"
	"nosql_db/internal/storage"
)

func handleDelete(req api.Request) api.Response {
	// Используем очередь для write-операции
	result := storage.GlobalManager.Enqueue(req.Database, func(coll *storage.Collection) (storage.WriteResult, error) {
		// Находим документы для удаления через FullScan
		allDocs := coll.All()
		deletedCount := 0

		for _, doc := range allDocs {
			if operators.MatchDocument(doc, req.Query) {
				if id, ok := doc["_id"].(string); ok {
					if coll.Delete(id) {
						deletedCount++
					}
				}
			}
		}

		if deletedCount > 0 {
			if err := coll.Save(); err != nil {
				return storage.WriteResult{}, fmt.Errorf("failed to save changes: %w", err)
			}
			if err := coll.RebuildAllIndexes(); err != nil {
				return storage.WriteResult{}, fmt.Errorf("failed to rebuild indexes: %w", err)
			}
		}

		return storage.WriteResult{
			DeletedCount: deletedCount,
			Message:      fmt.Sprintf("Deleted %d document(s)", deletedCount),
		}, nil
	})

	if result.Error != nil {
		return api.Response{Status: api.StatusError, Message: result.Error.Error()}
	}

	return api.Response{
		Status:  api.StatusSuccess,
		Message: result.Message,
		Count:   result.DeletedCount,
	}
}
