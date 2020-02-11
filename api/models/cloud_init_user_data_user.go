// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// CloudInitUserDataUser cloud init user data user
// swagger:model CloudInitUserDataUser
type CloudInitUserDataUser struct {

	// expire date
	ExpireDate string `json:"ExpireDate,omitempty"`

	// gecos
	Gecos string `json:"Gecos,omitempty"`

	// groups
	Groups []string `json:"Groups"`

	// inactive
	Inactive bool `json:"Inactive,omitempty"`

	// lock password
	LockPassword bool `json:"LockPassword,omitempty"`

	// name
	Name string `json:"Name,omitempty"`

	// password
	Password string `json:"Password,omitempty"`

	// primary group
	PrimaryGroup string `json:"PrimaryGroup,omitempty"`

	// shell
	Shell string `json:"Shell,omitempty"`

	// Ssh authorised keys
	SSHAuthorisedKeys []string `json:"SshAuthorisedKeys"`

	// sudo
	Sudo string `json:"Sudo,omitempty"`

	// system
	System bool `json:"System,omitempty"`
}

// Validate validates this cloud init user data user
func (m *CloudInitUserDataUser) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *CloudInitUserDataUser) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *CloudInitUserDataUser) UnmarshalBinary(b []byte) error {
	var res CloudInitUserDataUser
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}