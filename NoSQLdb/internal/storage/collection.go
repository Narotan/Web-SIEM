package storage

import (
	"fmt"
	"math/rand"
	"nosql_db/internal/index"
	"sync"
	"time"
)

type Collection struct {
	mutex   sync.RWMutex
	Name    string
	Data    *HashMap
	Indexes map[string]*index.BTree
}

func NewCollection(name string) *Collection {
	return &Collection{
		Name:    name,
		Data:    NewHashMap(),
		Indexes: make(map[string]*index.BTree),
	}
}

func generateID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(1000000))
}

func (c *Collection) Insert(doc map[string]any) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	id := generateID()
	doc["_id"] = id
	c.Data.Put(id, doc)

	c.updateIndexesOnInsert(id, doc)

	return id, nil
}

// GetByID получает документ по _id
func (c *Collection) GetByID(id string) (map[string]any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	val, ok := c.Data.Get(id)
	if !ok {
		return nil, false
	}

	doc, ok := val.(map[string]any)
	if !ok {
		return nil, false
	}
	return doc, true
}

// Delete удаляет документ по _id
func (c *Collection) Delete(id string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	val, ok := c.Data.Get(id)
	if !ok {
		return false
	}
	doc := val.(map[string]any)

	c.updateIndexesOnDelete(id, doc)

	return c.Data.Remove(id)
}

func (c *Collection) All() []map[string]any {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	items := c.Data.Items()
	docs := make([]map[string]any, 0, len(items))
	for _, v := range items {
		if doc, ok := v.(map[string]any); ok {
			docs = append(docs, doc)
		}
	}
	return docs
}
