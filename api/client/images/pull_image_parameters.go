// Code generated by go-swagger; DO NOT EDIT.

package images

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
)

// NewPullImageParams creates a new PullImageParams object
// with the default values initialized.
func NewPullImageParams() *PullImageParams {
	var ()
	return &PullImageParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewPullImageParamsWithTimeout creates a new PullImageParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewPullImageParamsWithTimeout(timeout time.Duration) *PullImageParams {
	var ()
	return &PullImageParams{

		timeout: timeout,
	}
}

// NewPullImageParamsWithContext creates a new PullImageParams object
// with the default values initialized, and the ability to set a context for a request
func NewPullImageParamsWithContext(ctx context.Context) *PullImageParams {
	var ()
	return &PullImageParams{

		Context: ctx,
	}
}

// NewPullImageParamsWithHTTPClient creates a new PullImageParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewPullImageParamsWithHTTPClient(client *http.Client) *PullImageParams {
	var ()
	return &PullImageParams{
		HTTPClient: client,
	}
}

/*PullImageParams contains all the parameters to send to the API endpoint
for the pull image operation typically these are written to a http.Request
*/
type PullImageParams struct {

	/*SourceURI
	  SourceURI for remote pull

	*/
	SourceURI *string
	/*TargetStorage
	  Storage Target

	*/
	TargetStorage *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the pull image params
func (o *PullImageParams) WithTimeout(timeout time.Duration) *PullImageParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the pull image params
func (o *PullImageParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the pull image params
func (o *PullImageParams) WithContext(ctx context.Context) *PullImageParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the pull image params
func (o *PullImageParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the pull image params
func (o *PullImageParams) WithHTTPClient(client *http.Client) *PullImageParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the pull image params
func (o *PullImageParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithSourceURI adds the sourceURI to the pull image params
func (o *PullImageParams) WithSourceURI(sourceURI *string) *PullImageParams {
	o.SetSourceURI(sourceURI)
	return o
}

// SetSourceURI adds the sourceUri to the pull image params
func (o *PullImageParams) SetSourceURI(sourceURI *string) {
	o.SourceURI = sourceURI
}

// WithTargetStorage adds the targetStorage to the pull image params
func (o *PullImageParams) WithTargetStorage(targetStorage *string) *PullImageParams {
	o.SetTargetStorage(targetStorage)
	return o
}

// SetTargetStorage adds the targetStorage to the pull image params
func (o *PullImageParams) SetTargetStorage(targetStorage *string) {
	o.TargetStorage = targetStorage
}

// WriteToRequest writes these params to a swagger request
func (o *PullImageParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.SourceURI != nil {

		// form param sourceURI
		var frSourceURI string
		if o.SourceURI != nil {
			frSourceURI = *o.SourceURI
		}
		fSourceURI := frSourceURI
		if fSourceURI != "" {
			if err := r.SetFormParam("sourceURI", fSourceURI); err != nil {
				return err
			}
		}

	}

	if o.TargetStorage != nil {

		// form param targetStorage
		var frTargetStorage string
		if o.TargetStorage != nil {
			frTargetStorage = *o.TargetStorage
		}
		fTargetStorage := frTargetStorage
		if fTargetStorage != "" {
			if err := r.SetFormParam("targetStorage", fTargetStorage); err != nil {
				return err
			}
		}

	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
