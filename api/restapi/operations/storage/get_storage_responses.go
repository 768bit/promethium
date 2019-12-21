// Code generated by go-swagger; DO NOT EDIT.

package storage

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/768bit/promethium/api/models"
)

// GetStorageOKCode is the HTTP code returned for type GetStorageOK
const GetStorageOKCode int = 200

/*GetStorageOK successful operation

swagger:response getStorageOK
*/
type GetStorageOK struct {

	/*
	  In: Body
	*/
	Payload models.Storage `json:"body,omitempty"`
}

// NewGetStorageOK creates GetStorageOK with default headers values
func NewGetStorageOK() *GetStorageOK {

	return &GetStorageOK{}
}

// WithPayload adds the payload to the get storage o k response
func (o *GetStorageOK) WithPayload(payload models.Storage) *GetStorageOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get storage o k response
func (o *GetStorageOK) SetPayload(payload models.Storage) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetStorageOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// GetStorageBadRequestCode is the HTTP code returned for type GetStorageBadRequest
const GetStorageBadRequestCode int = 400

/*GetStorageBadRequest Invalid ID supplied

swagger:response getStorageBadRequest
*/
type GetStorageBadRequest struct {
}

// NewGetStorageBadRequest creates GetStorageBadRequest with default headers values
func NewGetStorageBadRequest() *GetStorageBadRequest {

	return &GetStorageBadRequest{}
}

// WriteResponse to the client
func (o *GetStorageBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(400)
}

// GetStorageNotFoundCode is the HTTP code returned for type GetStorageNotFound
const GetStorageNotFoundCode int = 404

/*GetStorageNotFound Storage not found

swagger:response getStorageNotFound
*/
type GetStorageNotFound struct {
}

// NewGetStorageNotFound creates GetStorageNotFound with default headers values
func NewGetStorageNotFound() *GetStorageNotFound {

	return &GetStorageNotFound{}
}

// WriteResponse to the client
func (o *GetStorageNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(404)
}