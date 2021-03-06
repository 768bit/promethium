// Code generated by go-swagger; DO NOT EDIT.

package images

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	strfmt "github.com/go-openapi/strfmt"
)

// NewPullImageParams creates a new PullImageParams object
// no default values defined in spec.
func NewPullImageParams() PullImageParams {

	return PullImageParams{}
}

// PullImageParams contains all the bound params for the pull image operation
// typically these are obtained from a http.Request
//
// swagger:parameters pullImage
type PullImageParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*SourceURI for remote pull
	  In: formData
	*/
	SourceURI *string
	/*Storage Target
	  In: formData
	*/
	TargetStorage *string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewPullImageParams() beforehand.
func (o *PullImageParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if err != http.ErrNotMultipart {
			return errors.New(400, "%v", err)
		} else if err := r.ParseForm(); err != nil {
			return errors.New(400, "%v", err)
		}
	}
	fds := runtime.Values(r.Form)

	fdSourceURI, fdhkSourceURI, _ := fds.GetOK("sourceURI")
	if err := o.bindSourceURI(fdSourceURI, fdhkSourceURI, route.Formats); err != nil {
		res = append(res, err)
	}

	fdTargetStorage, fdhkTargetStorage, _ := fds.GetOK("targetStorage")
	if err := o.bindTargetStorage(fdTargetStorage, fdhkTargetStorage, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindSourceURI binds and validates parameter SourceURI from formData.
func (o *PullImageParams) bindSourceURI(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false

	if raw == "" { // empty values pass all other validations
		return nil
	}

	o.SourceURI = &raw

	return nil
}

// bindTargetStorage binds and validates parameter TargetStorage from formData.
func (o *PullImageParams) bindTargetStorage(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false

	if raw == "" { // empty values pass all other validations
		return nil
	}

	o.TargetStorage = &raw

	return nil
}
