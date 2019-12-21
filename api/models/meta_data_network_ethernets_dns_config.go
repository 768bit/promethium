// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"strconv"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// MetaDataNetworkEthernetsDNSConfig meta data network ethernets DNS config
// swagger:model MetaDataNetworkEthernetsDNSConfig
type MetaDataNetworkEthernetsDNSConfig struct {

	// addresses
	Addresses []string `json:"addresses"`

	// search
	Search []strfmt.Hostname `json:"search"`
}

// Validate validates this meta data network ethernets DNS config
func (m *MetaDataNetworkEthernetsDNSConfig) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateSearch(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *MetaDataNetworkEthernetsDNSConfig) validateSearch(formats strfmt.Registry) error {

	if swag.IsZero(m.Search) { // not required
		return nil
	}

	for i := 0; i < len(m.Search); i++ {

		if err := validate.FormatOf("search"+"."+strconv.Itoa(i), "body", "hostname", m.Search[i].String(), formats); err != nil {
			return err
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *MetaDataNetworkEthernetsDNSConfig) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *MetaDataNetworkEthernetsDNSConfig) UnmarshalBinary(b []byte) error {
	var res MetaDataNetworkEthernetsDNSConfig
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}