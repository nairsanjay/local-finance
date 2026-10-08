package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"local-finance/internal/db"
	"local-finance/internal/investment"
	"local-finance/internal/models"
)

const MaxInvestmentFileSize = 10 << 20

type InvestmentService struct {
	db       *db.DB
	registry *investment.Registry
}

func NewInvestmentService(database *db.DB) *InvestmentService {
	return &InvestmentService{db: database, registry: investment.DefaultRegistry}
}
func (s *InvestmentService) Import(filename string, r io.Reader) (*models.InvestmentSnapshot, bool, error) {
	snapshot, hash, err := s.parse(filename, r)
	if err != nil {
		return nil, false, err
	}
	snapshot.ID = uuid.NewString()
	snapshot.ImportedAt = time.Now().UTC().Format(time.RFC3339Nano)
	return s.db.SaveInvestmentSnapshot(snapshot, hash)
}

func (s *InvestmentService) Preview(filename string, r io.Reader) (*models.InvestmentSnapshot, error) {
	snapshot, _, err := s.parse(filename, r)
	return snapshot, err
}

func (s *InvestmentService) parse(filename string, r io.Reader) (*models.InvestmentSnapshot, string, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxInvestmentFileSize+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) > MaxInvestmentFileSize {
		return nil, "", fmt.Errorf("investment file exceeds 10 MB")
	}
	snapshot, err := s.registry.Parse(filename, data)
	if err != nil {
		return nil, "", err
	}
	snapshot.Filename = filepath.Base(filename)
	hash := sha256.Sum256(data)
	return snapshot, hex.EncodeToString(hash[:]), nil
}
