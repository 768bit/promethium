// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// UpdateVM update VM
// swagger:model UpdateVM
type UpdateVM struct {

	// auto start
	AutoStart bool `json:"autoStart,omitempty"`

	// boot cmd
	BootCmd string `json:"bootCmd,omitempty"`

	// cpus
	Cpus int64 `json:"cpus,omitempty"`

	// entry point
	EntryPoint string `json:"entryPoint,omitempty"`

	// memory
	Memory int64 `json:"memory,omitempty"`

	// name
	Name string `json:"name,omitempty"`
}

// Validate validates this update VM
func (m *UpdateVM) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *UpdateVM) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *UpdateVM) UnmarshalBinary(b []byte) error {
	var res UpdateVM
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}