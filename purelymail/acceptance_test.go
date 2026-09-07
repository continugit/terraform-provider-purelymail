package purelymail

import (
	"os"
	"testing"
)

func TestAccConfiguration(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("set TF_ACC=1 to run acceptance tests")
	}
	if os.Getenv("PURELYMAIL_API_TOKEN") == "" {
		t.Fatal("PURELYMAIL_API_TOKEN must be set for acceptance tests")
	}
	t.Skip("live acceptance scenarios are enabled after test fixtures are configured")
}
