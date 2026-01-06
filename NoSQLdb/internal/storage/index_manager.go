package storage

import (
	"encoding/json"
	"fmt"
	"nosql_db/internal/index"
	"os"
	"path/filepath"
)

// CreateIndex создает индекс на указанном поле
func (c *Collection) CreateIndex(fieldName string, order int) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.Indexes[fieldName]; exists {
		return fmt.Errorf("index on field '%s' already exists", fieldName)
	}
	btree := index.NewBPlusTree(order)

	items := c.Data.Items()
	for _, v := range items {
		doc, ok := v.(map[string]any)
		if !ok {
			continue
		}
		if fieldValue, exists := doc[fieldName]; exists {
			docID := doc["_id"].(string)
			key := index.ValueToKey(fieldValue)
			btree.Insert(key, []byte(docID))
		}
	}
	c.Indexes[fieldName] = btree

	return c.saveIndexInternal(fieldName)
}

// HasIndex проверяет существование индекса на поле
func (c *Collection) HasIndex(fieldName string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	_, exists := c.Indexes[fieldName]
	return exists
}

// GetIndex возвращает индекс для поля
func (c *Collection) GetIndex(fieldName string) (*index.BTree, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	btree, exists := c.Indexes[fieldName]
	return btree, exists
}

// LoadIndex загружает индекс с диска
func (c *Collection) LoadIndex(fieldName string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.loadIndexInternal(fieldName)
}

// loadIndexInternal - приватная версия без блокировок
func (c *Collection) loadIndexInternal(fieldName string) error {
	indexPath := filepath.Join("data", "indexes", fmt.Sprintf("%s_%s.idx", c.Name, fieldName))
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return nil
	}
	jsonData, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read index file: %w", err)
	}
	var indexData IndexFile
	if err := json.Unmarshal(jsonData, &indexData); err != nil {
		return fmt.Errorf("failed to unmarshal index: %w", err)
	}
	btree := deserializeBTree(&indexData)
	c.Indexes[fieldName] = btree
	return nil
}

// LoadAllIndexes загружает все индексы для коллекции
func (c *Collection) LoadAllIndexes() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	indexDir := filepath.Join("data", "indexes")
	if _, err := os.Stat(indexDir); os.IsNotExist(err) {
		return nil
	}
	entries, err := os.ReadDir(indexDir)
	if err != nil {
		return fmt.Errorf("failed to read index directory: %w", err)
	}
	prefix := c.Name + "_"
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) > len(prefix) && name[:len(prefix)] == prefix && filepath.Ext(name) == ".idx" {
			fieldName := name[len(prefix) : len(name)-4]
			if err := c.loadIndexInternal(fieldName); err != nil {
				return err
			}
		}
	}
	return nil
}

// SaveIndex сохраняет индекс на диск (Публичный метод)
func (c *Collection) SaveIndex(fieldName string) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.saveIndexInternal(fieldName)
}

// saveIndexInternal - сохранение без блокировок (для использования внутри CreateIndex)
func (c *Collection) saveIndexInternal(fieldName string) error {
	btree, exists := c.Indexes[fieldName]
	if !exists {
		return fmt.Errorf("index on field '%s' does not exist", fieldName)
	}
	indexPath := filepath.Join("data", "indexes", fmt.Sprintf("%s_%s.idx", c.Name, fieldName))
	if err := os.MkdirAll(filepath.Dir(indexPath), 0755); err != nil {
		return fmt.Errorf("failed to create index directory: %w", err)
	}
	indexData := serializeBTree(btree, fieldName, 64)
	jsonData, err := json.MarshalIndent(indexData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}
	if err := os.WriteFile(indexPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}
	return nil
}

// SaveAllIndexes сохраняет все индексы на диск
func (c *Collection) SaveAllIndexes() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for fieldName := range c.Indexes {
		if err := c.saveIndexInternal(fieldName); err != nil {
			return err
		}
	}
	return nil
}

// RebuildAllIndexes пересоздает все индексы
func (c *Collection) RebuildAllIndexes() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	fields := make([]string, 0, len(c.Indexes))
	for fieldName := range c.Indexes {
		fields = append(fields, fieldName)
	}
	c.Indexes = make(map[string]*index.BTree)

	items := c.Data.Items()

	for _, fieldName := range fields {
		btree := index.NewBPlusTree(64)
		for _, v := range items {
			doc, ok := v.(map[string]any)
			if !ok {
				continue
			}

			if fieldValue, exists := doc[fieldName]; exists {
				docID := doc["_id"].(string)
				key := index.ValueToKey(fieldValue)
				btree.Insert(key, []byte(docID))
			}
		}
		c.Indexes[fieldName] = btree
		if err := c.saveIndexInternal(fieldName); err != nil {
			return err
		}
	}
	return nil
}

// updateIndexesOnInsert (Приватный) - вызывается внутри Insert, мьютексы не нужны
func (c *Collection) updateIndexesOnInsert(docID string, doc map[string]any) {
	for fieldName, btree := range c.Indexes {
		if fieldValue, exists := doc[fieldName]; exists {
			key := index.ValueToKey(fieldValue)
			btree.Insert(key, []byte(docID))
		}
	}
}

// updateIndexesOnDelete (Приватный) - вызывается внутри Delete, мьютексы не нужны
func (c *Collection) updateIndexesOnDelete(docID string, doc map[string]any) {
	for fieldName, btree := range c.Indexes {
		if fieldValue, exists := doc[fieldName]; exists {
			key := index.ValueToKey(fieldValue)
			btree.Delete(key, []byte(docID))
		}
	}
}
