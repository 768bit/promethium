// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/validate"
)

// ImageContains image contains
// swagger:model ImageContains
type ImageContains string

const (

	// ImageContainsKernel captures enum value "Kernel"
	ImageContainsKernel ImageContains = "Kernel"

	// ImageContainsRootDisk captures enum value "RootDisk"
	ImageContainsRootDisk ImageContains = "RootDisk"

	// ImageContainsAdditionalDisks captures enum value "AdditionalDisks"
	ImageContainsAdditionalDisks ImageContains = "AdditionalDisks"

	// ImageContainsCloudInitUserData captures enum value "CloudInitUserData"
	ImageContainsCloudInitUserData ImageContains = "CloudInitUserData"
)

// for schema
var imageContainsEnum []interface{}

func init() {
	var res []ImageContains
	if err := json.Unmarshal([]byte(`["Kernel","RootDisk","AdditionalDisks","CloudInitUserData"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		imageContainsEnum = append(imageContainsEnum, v)
	}
}

func (m ImageContains) validateImageContainsEnum(path, location string, value ImageContains) error {
	if err := validate.Enum(path, location, value, imageContainsEnum); err != nil {
		return err
	}
	return nil
}

// Validate validates this image contains
func (m ImageContains) Validate(formats strfmt.Registry) error {
	var res []error

	// value enum
	if err := m.validateImageContainsEnum("", "body", m); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
