package purelymail

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientSendsAuthenticationAndDecodesResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Purelymail-Api-Token") != "secret" {
			t.Errorf("unexpected token header: %q", r.Header.Get("Purelymail-Api-Token"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"code":"ownership-code"}}`))
	}))
	defer server.Close()

	var result ownershipResponse
	if err := NewClient(server.URL, "secret").Do(context.Background(), "getOwnershipCode", map[string]any{}, &result); err != nil {
		t.Fatal(err)
	}
	if result.Code != "ownership-code" {
		t.Fatalf("unexpected result: %q", result.Code)
	}
}

func TestClientReturnsAPIErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"error":{"code":"bad_request","message":"invalid payload"}}`))
	}))
	defer server.Close()

	if err := NewClient(server.URL, "secret").Do(context.Background(), "createUser", map[string]any{}, &Empty{}); err == nil {
		t.Fatal("expected API error")
	}
}
