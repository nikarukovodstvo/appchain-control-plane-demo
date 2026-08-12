package configstore

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPCreateReplayAndRead(t *testing.T) {
	h := NewHandler(NewService())
	payload, _ := json.Marshal(validInput())
	request := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/v1/chain-configs", bytes.NewReader(payload))
		req.Header.Set("Idempotency-Key", "http-request-1")
		recorder := httptest.NewRecorder()
		h.ServeHTTP(recorder, req)
		return recorder
	}
	first := request()
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", first.Code, first.Body.String())
	}
	var created Config
	if err := json.NewDecoder(first.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	second := request()
	if second.Code != http.StatusOK || second.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatalf("expected replay response, got %d header=%q", second.Code, second.Header().Get("Idempotency-Replayed"))
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/chain-configs/"+created.ID, nil)
	getRecorder := httptest.NewRecorder()
	h.ServeHTTP(getRecorder, getReq)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("expected read success, got %d", getRecorder.Code)
	}
}

func TestHTTPRejectsUnknownFields(t *testing.T) {
	h := NewHandler(NewService())
	req := httptest.NewRequest(http.MethodPost, "/v1/chain-configs", bytes.NewBufferString(`{"name":"x","unknown":true}`))
	req.Header.Set("Idempotency-Key", "bad-request")
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

