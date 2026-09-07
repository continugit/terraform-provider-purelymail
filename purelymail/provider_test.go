package purelymail

import (
	"testing"
)

func TestProviderFactory(t *testing.T) {
	if New("test")() == nil {
		t.Fatal("provider factory returned nil")
	}
}
