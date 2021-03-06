// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// DiskImage disk image
// swagger:model DiskImage
type DiskImage struct {

	// hash
	Hash string `json:"Hash,omitempty"`

	// is root
	IsRoot bool `json:"IsRoot,omitempty"`

	// size
	Size int64 `json:"Size,omitempty"`
}

// Validate validates this disk image
func (m *DiskImage) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *DiskImage) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *DiskImage) UnmarshalBinary(b []byte) error {
	var res DiskImage
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
