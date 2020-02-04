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

// NewPushImageParams creates a new PushImageParams object
// with the default values initialized.
func NewPushImageParams() *PushImageParams {
	var ()
	return &PushImageParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewPushImageParamsWithTimeout creates a new PushImageParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewPushImageParamsWithTimeout(timeout time.Duration) *PushImageParams {
	var ()
	return &PushImageParams{

		timeout: timeout,
	}
}

// NewPushImageParamsWithContext creates a new PushImageParams object
// with the default values initialized, and the ability to set a context for a request
func NewPushImageParamsWithContext(ctx context.Context) *PushImageParams {
	var ()
	return &PushImageParams{

		Context: ctx,
	}
}

// NewPushImageParamsWithHTTPClient creates a new PushImageParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewPushImageParamsWithHTTPClient(client *http.Client) *PushImageParams {
	var ()
	return &PushImageParams{
		HTTPClient: client,
	}
}

/*PushImageParams contains all the parameters to send to the API endpoint
for the push image operation typically these are written to a http.Request
*/
type PushImageParams struct {

	/*InFileBlob
	  The file to upload.

	*/
	InFileBlob runtime.NamedReadCloser
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

// WithTimeout adds the timeout to the push image params
func (o *PushImageParams) WithTimeout(timeout time.Duration) *PushImageParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the push image params
func (o *PushImageParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the push image params
func (o *PushImageParams) WithContext(ctx context.Context) *PushImageParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the push image params
func (o *PushImageParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the push image params
func (o *PushImageParams) WithHTTPClient(client *http.Client) *PushImageParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the push image params
func (o *PushImageParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithInFileBlob adds the inFileBlob to the push image params
func (o *PushImageParams) WithInFileBlob(inFileBlob runtime.NamedReadCloser) *PushImageParams {
	o.SetInFileBlob(inFileBlob)
	return o
}

// SetInFileBlob adds the inFileBlob to the push image params
func (o *PushImageParams) SetInFileBlob(inFileBlob runtime.NamedReadCloser) {
	o.InFileBlob = inFileBlob
}

// WithSourceURI adds the sourceURI to the push image params
func (o *PushImageParams) WithSourceURI(sourceURI *string) *PushImageParams {
	o.SetSourceURI(sourceURI)
	return o
}

// SetSourceURI adds the sourceUri to the push image params
func (o *PushImageParams) SetSourceURI(sourceURI *string) {
	o.SourceURI = sourceURI
}

// WithTargetStorage adds the targetStorage to the push image params
func (o *PushImageParams) WithTargetStorage(targetStorage *string) *PushImageParams {
	o.SetTargetStorage(targetStorage)
	return o
}

// SetTargetStorage adds the targetStorage to the push image params
func (o *PushImageParams) SetTargetStorage(targetStorage *string) {
	o.TargetStorage = targetStorage
}

// WriteToRequest writes these params to a swagger request
func (o *PushImageParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.InFileBlob != nil {

		if o.InFileBlob != nil {

			// form file param inFileBlob
			if err := r.SetFileParam("inFileBlob", o.InFileBlob); err != nil {
				return err
			}

		}

	}

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
