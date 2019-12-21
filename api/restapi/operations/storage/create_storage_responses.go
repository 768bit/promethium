// Code generated by go-swagger; DO NOT EDIT.

package storage

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/768bit/promethium/api/models"
)

// CreateStorageOKCode is the HTTP code returned for type CreateStorageOK
const CreateStorageOKCode int = 200

/*CreateStorageOK successful operation

swagger:response createStorageOK
*/
type CreateStorageOK struct {

	/*
	  In: Body
	*/
	Payload models.Storage `json:"body,omitempty"`
}

// NewCreateStorageOK creates CreateStorageOK with default headers values
func NewCreateStorageOK() *CreateStorageOK {

	return &CreateStorageOK{}
}

// WithPayload adds the payload to the create storage o k response
func (o *CreateStorageOK) WithPayload(payload models.Storage) *CreateStorageOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create storage o k response
func (o *CreateStorageOK) SetPayload(payload models.Storage) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateStorageOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}