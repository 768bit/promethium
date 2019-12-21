// Code generated by go-swagger; DO NOT EDIT.

package vms

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/768bit/promethium/api/models"
)

// NewUpdateVMVolumeParams creates a new UpdateVMVolumeParams object
// no default values defined in spec.
func NewUpdateVMVolumeParams() UpdateVMVolumeParams {

	return UpdateVMVolumeParams{}
}

// UpdateVMVolumeParams contains all the bound params for the update VM volume operation
// typically these are obtained from a http.Request
//
// swagger:parameters updateVMVolume
type UpdateVMVolumeParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*ID of VM to return
	  Required: true
	  In: path
	*/
	VMID string
	/*Pet to add to the store
	  Required: true
	  In: body
	*/
	VMInterfaceConfig *models.UpdateVMVolume
	/*ID of VM Volume to use
	  Required: true
	  In: path
	*/
	VolumeID string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewUpdateVMVolumeParams() beforehand.
func (o *UpdateVMVolumeParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rVMID, rhkVMID, _ := route.Params.GetOK("vmID")
	if err := o.bindVMID(rVMID, rhkVMID, route.Formats); err != nil {
		res = append(res, err)
	}

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.UpdateVMVolume
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("vmInterfaceConfig", "body"))
			} else {
				res = append(res, errors.NewParseError("vmInterfaceConfig", "body", "", err))
			}
		} else {
			// validate body object
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

			if len(res) == 0 {
				o.VMInterfaceConfig = &body
			}
		}
	} else {
		res = append(res, errors.Required("vmInterfaceConfig", "body"))
	}
	rVolumeID, rhkVolumeID, _ := route.Params.GetOK("volumeID")
	if err := o.bindVolumeID(rVolumeID, rhkVolumeID, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindVMID binds and validates parameter VMID from path.
func (o *UpdateVMVolumeParams) bindVMID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	o.VMID = raw

	return nil
}

// bindVolumeID binds and validates parameter VolumeID from path.
func (o *UpdateVMVolumeParams) bindVolumeID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	o.VolumeID = raw

	return nil
}