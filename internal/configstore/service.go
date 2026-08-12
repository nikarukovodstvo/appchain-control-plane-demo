package configstore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidInput        = errors.New("invalid chain configuration")
	ErrIdempotencyConflict = errors.New("idempotency key reused with different input")
	ErrNotFound            = errors.New("chain configuration not found")
)

type Input struct {
	Name          string `json:"name"`
	ChainID       uint64 `json:"chainId"`
	RPCURL        string `json:"rpcUrl"`
	BlockTimeMS   uint64 `json:"blockTimeMs"`
	SequencerMode string `json:"sequencerMode"`
}

type Config struct {
	ID string `json:"id"`
	Input
	CreatedAt time.Time `json:"createdAt"`
}

type idempotencyRecord struct {
	inputHash string
	configID  string
}

type Service struct {
	mu          sync.RWMutex
	configs     map[string]Config
	idempotency map[string]idempotencyRecord
	now         func() time.Time
}

func NewService() *Service {
	return &Service{
		configs:     make(map[string]Config),
		idempotency: make(map[string]idempotencyRecord),
		now:         time.Now,
	}
}

func (s *Service) Create(key string, input Input) (Config, bool, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return Config{}, false, fmt.Errorf("%w: Idempotency-Key is required", ErrInvalidInput)
	}
	if err := validate(input); err != nil {
		return Config{}, false, err
	}

	payload, _ := json.Marshal(input)
	digest := sha256.Sum256(payload)
	inputHash := hex.EncodeToString(digest[:])

	s.mu.Lock()
	defer s.mu.Unlock()
	if previous, ok := s.idempotency[key]; ok {
		if previous.inputHash != inputHash {
			return Config{}, false, ErrIdempotencyConflict
		}
		return s.configs[previous.configID], true, nil
	}

	idDigest := sha256.Sum256([]byte(key + ":" + inputHash))
	id := "cfg_" + hex.EncodeToString(idDigest[:8])
	created := Config{ID: id, Input: input, CreatedAt: s.now().UTC()}
	s.configs[id] = created
	s.idempotency[key] = idempotencyRecord{inputHash: inputHash, configID: id}
	return created, false, nil
}

func (s *Service) Get(id string) (Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	config, ok := s.configs[id]
	if !ok {
		return Config{}, ErrNotFound
	}
	return config, nil
}

func validate(input Input) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if input.ChainID == 0 {
		return fmt.Errorf("%w: chainId must be positive", ErrInvalidInput)
	}
	parsed, err := url.ParseRequestURI(input.RPCURL)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
		return fmt.Errorf("%w: rpcUrl must be an absolute HTTP(S) URL", ErrInvalidInput)
	}
	if input.BlockTimeMS < 100 || input.BlockTimeMS > 120000 {
		return fmt.Errorf("%w: blockTimeMs must be between 100 and 120000", ErrInvalidInput)
	}
	if input.SequencerMode != "single" && input.SequencerMode != "committee" {
		return fmt.Errorf("%w: sequencerMode must be single or committee", ErrInvalidInput)
	}
	return nil
}

