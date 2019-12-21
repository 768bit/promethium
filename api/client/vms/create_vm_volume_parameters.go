// Code generated by go-swagger; DO NOT EDIT.

package vms

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/768bit/promethium/api/models"
)

// NewCreateVMVolumeParams creates a new CreateVMVolumeParams object
// with the default values initialized.
func NewCreateVMVolumeParams() *CreateVMVolumeParams {
	var ()
	return &CreateVMVolumeParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewCreateVMVolumeParamsWithTimeout creates a new CreateVMVolumeParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewCreateVMVolumeParamsWithTimeout(timeout time.Duration) *CreateVMVolumeParams {
	var ()
	return &CreateVMVolumeParams{

		timeout: timeout,
	}
}

// NewCreateVMVolumeParamsWithContext creates a new CreateVMVolumeParams object
// with the default values initialized, and the ability to set a context for a request
func NewCreateVMVolumeParamsWithContext(ctx context.Context) *CreateVMVolumeParams {
	var ()
	return &CreateVMVolumeParams{

		Context: ctx,
	}
}

// NewCreateVMVolumeParamsWithHTTPClient creates a new CreateVMVolumeParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewCreateVMVolumeParamsWithHTTPClient(client *http.Client) *CreateVMVolumeParams {
	var ()
	return &CreateVMVolumeParams{
		HTTPClient: client,
	}
}

/*CreateVMVolumeParams contains all the parameters to send to the API endpoint
for the create VM volume operation typically these are written to a http.Request
*/
type CreateVMVolumeParams struct {

	/*VMID
	  ID of VM to use

	*/
	VMID string
	/*VolumeConfig
	  VM Volume Config

	*/
	VolumeConfig *models.NewVMVolume

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the create VM volume params
func (o *CreateVMVolumeParams) WithTimeout(timeout time.Duration) *CreateVMVolumeParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the create VM volume params
func (o *CreateVMVolumeParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the create VM volume params
func (o *CreateVMVolumeParams) WithContext(ctx context.Context) *CreateVMVolumeParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the create VM volume params
func (o *CreateVMVolumeParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the create VM volume params
func (o *CreateVMVolumeParams) WithHTTPClient(client *http.Client) *CreateVMVolumeParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the create VM volume params
func (o *CreateVMVolumeParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithVMID adds the vMID to the create VM volume params
func (o *CreateVMVolumeParams) WithVMID(vMID string) *CreateVMVolumeParams {
	o.SetVMID(vMID)
	return o
}

// SetVMID adds the vmId to the create VM volume params
func (o *CreateVMVolumeParams) SetVMID(vMID string) {
	o.VMID = vMID
}

// WithVolumeConfig adds the volumeConfig to the create VM volume params
func (o *CreateVMVolumeParams) WithVolumeConfig(volumeConfig *models.NewVMVolume) *CreateVMVolumeParams {
	o.SetVolumeConfig(volumeConfig)
	return o
}

// SetVolumeConfig adds the volumeConfig to the create VM volume params
func (o *CreateVMVolumeParams) SetVolumeConfig(volumeConfig *models.NewVMVolume) {
	o.VolumeConfig = volumeConfig
}

// WriteToRequest writes these params to a swagger request
func (o *CreateVMVolumeParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param vmID
	if err := r.SetPathParam("vmID", o.VMID); err != nil {
		return err
	}

	if o.VolumeConfig != nil {
		if err := r.SetBodyParam(o.VolumeConfig); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}