package sender

import (
	"log"
	"sync"
	"time"

	"github.com/Narotan/SIEM-Agent/internal/domain"
	"github.com/Narotan/SIEM-Agent/internal/filter"
	"github.com/Narotan/SIEM-Agent/internal/storage"
)

type pipeline struct {
	sender     Sender
	config     Config
	diskBuffer *storage.DiskBuffer
	filter     filter.Filter

	buffer []domain.Event
	mu     sync.Mutex

	stop chan struct{}
	wg   sync.WaitGroup
}

func NewPipeline(sender Sender, cfg Config) Pipeline {
	p := &pipeline{
		sender:     sender,
		config:     cfg,
		diskBuffer: cfg.DiskBuffer,
		filter:     cfg.Filter,
		buffer:     make([]domain.Event, 0, cfg.BatchSize),
		stop:       make(chan struct{}),
	}

	if p.filter == nil {
		p.filter = filter.NewNoOpFilter()
	}

	return p
}

func (p *pipeline) Start(input <-chan domain.Event) {
	if p.diskBuffer != nil {
		go p.loadPendingBatches()
	}

	p.wg.Add(1)
	go p.run(input)
}

func (p *pipeline) loadPendingBatches() {
	batches, err := p.diskBuffer.LoadPending()
	if err != nil {
		log.Printf("Failed to load pending batches: %v", err)
		return
	}

	if len(batches) == 0 {
		return
	}

	log.Printf("Found %d pending batches on disk, sending...", len(batches))

	for _, batch := range batches {
		if err := p.sender.Send(batch); err != nil {
			log.Printf("Failed to send pending batch: %v", err)
			continue
		}

		if err := p.diskBuffer.Remove(batch); err != nil {
			log.Printf("Failed to remove batch from disk: %v", err)
		}
	}

	log.Printf("Successfully sent %d pending batches", len(batches))
}

func (p *pipeline) Stop() {
	close(p.stop)
	p.wg.Wait()

	p.mu.Lock()
	if len(p.buffer) > 0 {
		p.flush()
	}
	p.mu.Unlock()

	log.Println("Pipeline stopped gracefully")
}

func (p *pipeline) run(input <-chan domain.Event) {
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.FlushTimeout)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-input:
			if !ok {
				return
			}

			if !p.filter.Match(event) {
				continue
			}

			p.mu.Lock()
			p.buffer = append(p.buffer, event)

			if len(p.buffer) >= p.config.BatchSize {
				p.flush()
				ticker.Reset(p.config.FlushTimeout)
			}
			p.mu.Unlock()

		case <-ticker.C:
			p.mu.Lock()
			if len(p.buffer) > 0 {
				p.flush()
			}
			p.mu.Unlock()

		case <-p.stop:
			return
		}
	}
}

func (p *pipeline) flush() {
	if len(p.buffer) == 0 {
		return
	}

	batch := domain.Batch{
		AgentID:   p.config.AgentID,
		Timestamp: time.Now(),
		Events:    p.buffer,
	}

	if err := p.sender.Send(batch); err != nil {
		log.Printf("Failed to send batch: %v", err)

		if p.diskBuffer != nil {
			if saveErr := p.diskBuffer.Save(batch); saveErr != nil {
				log.Printf("CRITICAL: Failed to save batch to disk: %v", saveErr)
			} else {
				log.Printf("Batch saved to disk for later retry")
			}
		}
	}

	p.buffer = make([]domain.Event, 0, p.config.BatchSize)
}
