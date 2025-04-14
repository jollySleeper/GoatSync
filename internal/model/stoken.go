// Package model defines the GORM models for the GoatSync database schema.
// These models match the Django models in the original EteSync server.
package model

import (
	"goatsync/internal/crypto"

	"gorm.io/gorm"
)

// Stoken represents a sync token used for incremental synchronization.
// Every modification to collections/items creates a new Stoken.
// Clients send their last stoken to get only changes since that point.
//
// Django model reference:
//
//	class Stoken(models.Model):
//	    uid = models.CharField(db_index=True, unique=True, max_length=43)
type Stoken struct {
	ID  uint   `gorm:"primaryKey"`
	UID string `gorm:"uniqueIndex;size:43;not null"`
}

// TableName specifies the table name for GORM
func (Stoken) TableName() string {
	return "django_stoken"
}

// BeforeCreate generates a random UID before inserting a new Stoken
func (s *Stoken) BeforeCreate(tx *gorm.DB) error {
	if s.UID == "" {
		uid, err := crypto.GenerateStokenUID()
		if err != nil {
			return err
		}
		s.UID = uid
	}
	return nil
}

