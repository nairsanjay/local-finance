// Package investment contains investment statement adapters, independent of bank imports.
package investment

import (
	"fmt"
	"local-finance/internal/models"
)

type Parser interface {
	ID() string
	CanParse(filename string, data []byte) bool
	Parse(data []byte) (*models.InvestmentSnapshot, error)
}

type Registry struct{ parsers []Parser }

func NewRegistry() *Registry          { return &Registry{} }
func (r *Registry) Register(p Parser) { r.parsers = append(r.parsers, p) }
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

var DefaultRegistry = func() *Registry {
	r := NewRegistry()
	r.Register(ZerodhaHoldingsParser{})
	return r
}()
