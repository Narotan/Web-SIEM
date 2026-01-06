package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/Narotan/SIEM-Agent/internal/domain"
)

// DiskBuffer сохраняет батчи на диск при переполнении памяти
type DiskBuffer struct {
	directory   string
	maxSize     int64 // максимальный размер в байтах
	currentSize int64
	mu          sync.Mutex
}

// NewDiskBuffer создаёт новый дисковый буфер
func NewDiskBuffer(directory string, maxSize int64) (*DiskBuffer, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create buffer directory: %w", err)
	}

	db := &DiskBuffer{
		directory: directory,
		maxSize:   maxSize,
	}

	// Вычисляем текущий размер
	if err := db.calculateSize(); err != nil {
		return nil, err
	}

	return db, nil
}

// Save сохраняет батч на диск
func (db *DiskBuffer) Save(batch domain.Batch) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Проверяем не переполнен ли буфер
	if db.currentSize >= db.maxSize {
		return fmt.Errorf("disk buffer is full (%d bytes)", db.currentSize)
	}

	// Генерируем имя файла с timestamp
	filename := fmt.Sprintf("batch_%d.json", time.Now().UnixNano())
	path := filepath.Join(db.directory, filename)

	// Сериализуем батч
	data, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("failed to marshal batch: %w", err)
	}

	// Сохраняем в файл
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write batch file: %w", err)
	}

	db.currentSize += int64(len(data))
	return nil
}

// LoadPending загружает все несохранённые батчи с диска
func (db *DiskBuffer) LoadPending() ([]domain.Batch, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	files, err := os.ReadDir(db.directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read buffer directory: %w", err)
	}

	var batches []domain.Batch

	// Сортируем файлы по времени создания
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		path := filepath.Join(db.directory, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read batch file %s: %w", file.Name(), err)
		}

		var batch domain.Batch
		if err := json.Unmarshal(data, &batch); err != nil {
			return nil, fmt.Errorf("failed to unmarshal batch file %s: %w", file.Name(), err)
		}

		batches = append(batches, batch)
	}

	return batches, nil
}

// Remove удаляет файл батча после успешной отправки
func (db *DiskBuffer) Remove(batch domain.Batch) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Ищем файл с соответствующим timestamp
	files, err := os.ReadDir(db.directory)
	if err != nil {
		return fmt.Errorf("failed to read buffer directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(db.directory, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var savedBatch domain.Batch
		if err := json.Unmarshal(data, &savedBatch); err != nil {
			continue
		}

		// Сравниваем по AgentID и timestamp
		if savedBatch.AgentID == batch.AgentID &&
			savedBatch.Timestamp.Equal(batch.Timestamp) {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove batch file: %w", err)
			}
			db.currentSize -= int64(len(data))
			return nil
		}
	}

	return nil
}

// Clear очищает весь буфер на диске
func (db *DiskBuffer) Clear() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	files, err := os.ReadDir(db.directory)
	if err != nil {
		return fmt.Errorf("failed to read buffer directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(db.directory, file.Name())
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove file %s: %w", file.Name(), err)
		}
	}

	db.currentSize = 0
	return nil
}

// Size возвращает текущий размер буфера в байтах
func (db *DiskBuffer) Size() int64 {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.currentSize
}

// Count возвращает количество батчей на диске
func (db *DiskBuffer) Count() (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	files, err := os.ReadDir(db.directory)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			count++
		}
	}

	return count, nil
}

// calculateSize вычисляет текущий размер буфера
func (db *DiskBuffer) calculateSize() error {
	files, err := os.ReadDir(db.directory)
	if err != nil {
		return err
	}

	var totalSize int64
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		totalSize += info.Size()
	}

	db.currentSize = totalSize
	return nil
}

// Cleanup удаляет старые файлы, если буфер переполнен
func (db *DiskBuffer) Cleanup(maxAge time.Duration) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	files, err := os.ReadDir(db.directory)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// Удаляем файлы старше maxAge
		if now.Sub(info.ModTime()) > maxAge {
			path := filepath.Join(db.directory, file.Name())
			if err := os.Remove(path); err != nil {
				return err
			}
			db.currentSize -= info.Size()
		}
	}

	return nil
}

// HasSpace проверяет есть ли место для сохранения батча
func (db *DiskBuffer) HasSpace(estimatedSize int64) bool {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.currentSize+estimatedSize < db.maxSize
}

// SaveStream сохраняет батч используя streaming для больших данных
func (db *DiskBuffer) SaveStream(batch domain.Batch) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	filename := fmt.Sprintf("batch_%d.json", time.Now().UnixNano())
	path := filepath.Join(db.directory, filename)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create batch file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(batch); err != nil {
		os.Remove(path)
		return fmt.Errorf("failed to encode batch: %w", err)
	}

	info, err := file.Stat()
	if err != nil {
		return err
	}

	db.currentSize += info.Size()
	return nil
}

// LoadStream загружает батч используя streaming
func (db *DiskBuffer) LoadStream(filename string) (domain.Batch, error) {
	path := filepath.Join(db.directory, filename)

	file, err := os.Open(path)
	if err != nil {
		return domain.Batch{}, fmt.Errorf("failed to open batch file: %w", err)
	}
	defer file.Close()

	var batch domain.Batch
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&batch); err != nil {
		if err == io.EOF {
			return domain.Batch{}, fmt.Errorf("empty batch file")
		}
		return domain.Batch{}, fmt.Errorf("failed to decode batch: %w", err)
	}

	return batch, nil
}
