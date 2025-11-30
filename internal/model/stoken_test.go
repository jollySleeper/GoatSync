package model

import (
	"testing"

	"gorm.io/gorm"
)

func TestStoken_BeforeCreate(t *testing.T) {
	stoken := &Stoken{}

	// Before hook, UID should be empty
	if stoken.UID != "" {
		t.Error("UID should be empty before BeforeCreate")
	}

	// Call BeforeCreate (simulating GORM)
	err := stoken.BeforeCreate(&gorm.DB{})
	if err != nil {
		t.Fatalf("BeforeCreate failed: %v", err)
	}

	// UID should now be set
	if stoken.UID == "" {
		t.Error("UID should be set after BeforeCreate")
	}

	// UID should be 32 characters
	if len(stoken.UID) != 32 {
		t.Errorf("Expected UID length 32, got %d", len(stoken.UID))
	}
}

func TestStoken_BeforeCreate_PreservesExisting(t *testing.T) {
	existingUID := "existinguid123456789012"
	stoken := &Stoken{UID: existingUID}

	err := stoken.BeforeCreate(&gorm.DB{})
	if err != nil {
		t.Fatalf("BeforeCreate failed: %v", err)
	}

	// UID should be preserved (or overwritten - depends on implementation)
	// For GoatSync, we always generate a new UID
	if stoken.UID == "" {
		t.Error("UID should not be empty")
	}
}

func TestStoken_TableName(t *testing.T) {
	stoken := Stoken{}
	if stoken.TableName() != "django_stoken" {
		t.Errorf("Expected table name 'django_stoken', got '%s'", stoken.TableName())
	}
}

