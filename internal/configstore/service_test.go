package configstore

import (
	"errors"
	"sync"
	"testing"
)

func validInput() Input {
	return Input{Name: "orders-l2", ChainID: 90101, RPCURL: "https://rpc.example.test", BlockTimeMS: 2000, SequencerMode: "committee"}
}

func TestCreateIsIdempotent(t *testing.T) {
	svc := NewService()
	first, replayed, err := svc.Create("request-1", validInput())
	if err != nil || replayed {
		t.Fatalf("first create: replayed=%v err=%v", replayed, err)
	}
	second, replayed, err := svc.Create("request-1", validInput())
	if err != nil || !replayed || first.ID != second.ID {
		t.Fatalf("replay mismatch: first=%s second=%s replayed=%v err=%v", first.ID, second.ID, replayed, err)
	}
}

func TestCreateRejectsKeyReuseWithDifferentPayload(t *testing.T) {
	svc := NewService()
	_, _, _ = svc.Create("request-1", validInput())
	changed := validInput()
	changed.ChainID = 90102
	_, _, err := svc.Create("request-1", changed)
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestCreateIsConcurrentSafe(t *testing.T) {
	svc := NewService()
	var wg sync.WaitGroup
	ids := make(chan string, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg, _, err := svc.Create("same-request", validInput())
			if err != nil {
				t.Errorf("create failed: %v", err)
				return
			}
			ids <- cfg.ID
		}()
	}
	wg.Wait()
	close(ids)
	var first string
	for id := range ids {
		if first == "" {
			first = id
		}
		if id != first {
			t.Fatalf("expected one id, got %s and %s", first, id)
		}
	}
}

func TestValidation(t *testing.T) {
	tests := []Input{
		{},
		{Name: "x", ChainID: 1, RPCURL: "file:///tmp/rpc", BlockTimeMS: 1000, SequencerMode: "single"},
		{Name: "x", ChainID: 1, RPCURL: "https://rpc.example", BlockTimeMS: 10, SequencerMode: "single"},
		{Name: "x", ChainID: 1, RPCURL: "https://rpc.example", BlockTimeMS: 1000, SequencerMode: "unknown"},
	}
	for i, input := range tests {
		if _, _, err := NewService().Create("key", input); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("case %d expected invalid input, got %v", i, err)
		}
	}
}

