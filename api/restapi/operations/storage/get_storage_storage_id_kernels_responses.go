// Code generated by go-swagger; DO NOT EDIT.

package storage

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/768bit/promethium/api/models"
)

// GetStorageStorageIDKernelsOKCode is the HTTP code returned for type GetStorageStorageIDKernelsOK
const GetStorageStorageIDKernelsOKCode int = 200

/*GetStorageStorageIDKernelsOK OK

swagger:response getStorageStorageIdKernelsOK
*/
type GetStorageStorageIDKernelsOK struct {

	/*
	  In: Body
	*/
	Payload []models.StorageKernel `json:"body,omitempty"`
}

// NewGetStorageStorageIDKernelsOK creates GetStorageStorageIDKernelsOK with default headers values
func NewGetStorageStorageIDKernelsOK() *GetStorageStorageIDKernelsOK {

	return &GetStorageStorageIDKernelsOK{}
}

// WithPayload adds the payload to the get storage storage Id kernels o k response
func (o *GetStorageStorageIDKernelsOK) WithPayload(payload []models.StorageKernel) *GetStorageStorageIDKernelsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get storage storage Id kernels o k response
func (o *GetStorageStorageIDKernelsOK) SetPayload(payload []models.StorageKernel) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetStorageStorageIDKernelsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty array
		payload = make([]models.StorageKernel, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}