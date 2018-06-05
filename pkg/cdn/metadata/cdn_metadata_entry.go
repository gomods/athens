package metadata

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

// CdnMetadataEntry stores the module name and cdn URL.
type CdnMetadataEntry struct {
	ID          uuid.UUID `json:"id" db:"id" bson:"id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at" bson:"updated_at"`
	Module      string    `json:"module" db:"module" bson:"module"`
	RedirectURL string    `json:"redirect_url" db:"redirect_url" bson:"redirect_url"`
}

// String is not required by pop and may be deleted
func (e CdnMetadataEntry) String() string {
	je, _ := json.Marshal(e)
	return string(je)
}

// CdnMetadataEntries is not required by pop and may be deleted
type CdnMetadataEntries []CdnMetadataEntry

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (e *CdnMetadataEntry) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: e.Module, Name: "Module"},
		&validators.StringIsPresent{Field: e.RedirectURL, Name: "RedirectURL"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (e *CdnMetadataEntry) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (e *CdnMetadataEntry) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
