// Code generated by go-swagger; DO NOT EDIT.

package storage

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
)

// DestroyStorageOKCode is the HTTP code returned for type DestroyStorageOK
const DestroyStorageOKCode int = 200

/*DestroyStorageOK OK

swagger:response destroyStorageOK
*/
type DestroyStorageOK struct {
}

// NewDestroyStorageOK creates DestroyStorageOK with default headers values
func NewDestroyStorageOK() *DestroyStorageOK {

	return &DestroyStorageOK{}
}

// WriteResponse to the client
func (o *DestroyStorageOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}

// DestroyStorageBadRequestCode is the HTTP code returned for type DestroyStorageBadRequest
const DestroyStorageBadRequestCode int = 400

/*DestroyStorageBadRequest Invalid ID supplied

swagger:response destroyStorageBadRequest
*/
type DestroyStorageBadRequest struct {
}

// NewDestroyStorageBadRequest creates DestroyStorageBadRequest with default headers values
func NewDestroyStorageBadRequest() *DestroyStorageBadRequest {

	return &DestroyStorageBadRequest{}
}

// WriteResponse to the client
func (o *DestroyStorageBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(400)
}

// DestroyStorageNotFoundCode is the HTTP code returned for type DestroyStorageNotFound
const DestroyStorageNotFoundCode int = 404

/*DestroyStorageNotFound Storage not found

swagger:response destroyStorageNotFound
*/
type DestroyStorageNotFound struct {
}

// NewDestroyStorageNotFound creates DestroyStorageNotFound with default headers values
func NewDestroyStorageNotFound() *DestroyStorageNotFound {

	return &DestroyStorageNotFound{}
}

// WriteResponse to the client
func (o *DestroyStorageNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(404)
}