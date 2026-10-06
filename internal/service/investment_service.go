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
	data, err := io.ReadAll(io.LimitReader(r, MaxInvestmentFileSize+1))
	if err != nil {
		return nil, false, err
	}
	if len(data) > MaxInvestmentFileSize {
		return nil, false, fmt.Errorf("investment file exceeds 10 MB")
	}
	snapshot, err := s.registry.Parse(filename, data)
	if err != nil {
		return nil, false, err
	}
	snapshot.ID = uuid.NewString()
	snapshot.Filename = filepath.Base(filename)
	snapshot.ImportedAt = time.Now().UTC().Format(time.RFC3339Nano)
	hash := sha256.Sum256(data)
	return s.db.SaveInvestmentSnapshot(snapshot, hex.EncodeToString(hash[:]))
}
