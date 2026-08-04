package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseUsesFrontendSuccessCode(t *testing.T) {
	recorder := httptest.NewRecorder()

	Response(recorder, map[string]bool{"flag": true}, nil)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected HTTP status: %d", recorder.Code)
	}

	var body Body
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != http.StatusOK {
		t.Fatalf("expected application success code %d, got %d", http.StatusOK, body.Code)
	}
	if body.Msg != "OK" {
		t.Fatalf("unexpected success message: %q", body.Msg)
	}
}

func TestResponseKeepsErrorCode(t *testing.T) {
	recorder := httptest.NewRecorder()

	Response(recorder, nil, errors.New("boom"))

	var body Body
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != -1 {
		t.Fatalf("expected application error code -1, got %d", body.Code)
	}
	if body.Msg != "boom" {
		t.Fatalf("unexpected error message: %q", body.Msg)
	}
}
