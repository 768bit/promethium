// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// NewVM new VM
// swagger:model NewVM
type NewVM struct {

	// auto start
	AutoStart bool `json:"autoStart,omitempty"`

	// cpus
	Cpus int64 `json:"cpus,omitempty"`

	// from image
	FromImage string `json:"fromImage,omitempty"`

	// kernel image
	KernelImage string `json:"kernelImage,omitempty"`

	// memory
	Memory int64 `json:"memory,omitempty"`

	// name
	Name string `json:"name,omitempty"`

	// primary network ID
	PrimaryNetworkID string `json:"primaryNetworkID,omitempty"`

	// root disk size
	RootDiskSize int64 `json:"rootDiskSize,omitempty"`

	// storage name
	StorageName string `json:"storageName,omitempty"`
}

// Validate validates this new VM
func (m *NewVM) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *NewVM) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *NewVM) UnmarshalBinary(b []byte) error {
	var res NewVM
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
