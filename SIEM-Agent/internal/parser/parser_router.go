package parser

import (
	"log"
	"os"
	"sync"

	"github.com/Narotan/SIEM-Agent/internal/domain"
	"github.com/Narotan/SIEM-Agent/internal/reader"
)

type Parser interface {
	Match(event reader.RawEvent) bool
	Parse(event reader.RawEvent) (domain.Event, error)
}

// Router маршрутизирует события кратко
type Router struct {
	parsers  []Parser
	hostname string
	wg       sync.WaitGroup
}

// NewRouter создаёт маршрутизатор
func NewRouter(parsers ...Parser) *Router {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	return &Router{
		parsers:  parsers,
		hostname: hostname,
	}
}

// Start запускает обработку
func (r *Router) Start(rawEvents <-chan reader.RawEvent, errors <-chan error, output chan<- domain.Event) {
	r.wg.Add(2)

	go func() {
		defer r.wg.Done()
		for rawEvent := range rawEvents {
			event, err := r.route(rawEvent)
			if err != nil {
				log.Printf("parser-router: failed to parse event: %v", err)
				continue
			}

			event.Hostname = r.hostname

			output <- event
		}
	}()

	go func() {
		defer r.wg.Done()
		for err := range errors {
			log.Printf("parser-router: reader error: %v", err)
		}
	}()
}

func (r *Router) Stop() {
	r.wg.Wait()
}

func (r *Router) route(event reader.RawEvent) (domain.Event, error) {
	for _, p := range r.parsers {
		if p.Match(event) {
			return p.Parse(event)
		}
	}

	return domain.Event{
		Timestamp: event.Timestamp,
		Source:    "unknown",
		RawLog:    event.Data,
		EventType: "unparsed",
		Severity:  "info",
	}, nil
}
