// Package investment contains investment statement adapters, independent of bank imports.
package investment

import (
	"fmt"
	"local-finance/internal/models"
)

// Parser adapts one provider's export into the shared portfolio model.
// Adapters register themselves with DefaultRegistry from their init function.
type Parser interface {
	ID() string
	Info() ParserInfo
	CanParse(filename string, data []byte) bool
	Parse(data []byte) (*models.InvestmentSnapshot, error)
}

// ParserInfo describes the adapter's import capabilities to any client.
type ParserInfo struct {
	ID         string   `json:"id"`
	Provider   string   `json:"provider"`
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"`
}

type Registry struct{ parsers []Parser }

func NewRegistry() *Registry { return &Registry{} }

// Register replaces an adapter with the same stable ID, as the bank registry does.
func (r *Registry) Register(p Parser) {
	for i, existing := range r.parsers {
		if existing.ID() == p.ID() {
			r.parsers[i] = p
			return
		}
	}
	r.parsers = append(r.parsers, p)
}

func (r *Registry) Get(id string) (Parser, bool) {
	for _, p := range r.parsers {
		if p.ID() == id {
			return p, true
		}
	}
	return nil, false
}
func (r *Registry) List() []ParserInfo {
	result := make([]ParserInfo, 0, len(r.parsers))
	for _, p := range r.parsers {
		info := p.Info()
		info.ID = p.ID()
		info.Extensions = append([]string{}, info.Extensions...)
		result = append(result, info)
	}
	return result
}
func (r *Registry) Parse(filename string, data []byte) (*models.InvestmentSnapshot, error) {
	var selected Parser
	for _, p := range r.parsers {
		if p.CanParse(filename, data) {
			if selected != nil {
				return nil, fmt.Errorf("ambiguous investment statement format")
			}
			selected = p
		}
	}
	if selected == nil {
		return nil, fmt.Errorf("unsupported investment statement format")
	}
	result, err := selected.Parse(data)
	if err != nil {
		return nil, err
	}
	if result == nil || result.Provider == "" || result.AccountRef == "" || result.AsOf == "" || result.Currency == "" {
		return nil, fmt.Errorf("investment parser returned an incomplete portfolio")
	}
	result.ParserID = selected.ID()
	return result, nil
}

var DefaultRegistry = NewRegistry()
