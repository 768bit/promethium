// Code generated by go-swagger; DO NOT EDIT.

package vms

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/768bit/promethium/api/models"
)

// GetVMOKCode is the HTTP code returned for type GetVMOK
const GetVMOKCode int = 200

/*GetVMOK successful operation

swagger:response getVmOK
*/
type GetVMOK struct {

	/*
	  In: Body
	*/
	Payload *models.VM `json:"body,omitempty"`
}

// NewGetVMOK creates GetVMOK with default headers values
func NewGetVMOK() *GetVMOK {

	return &GetVMOK{}
}

// WithPayload adds the payload to the get Vm o k response
func (o *GetVMOK) WithPayload(payload *models.VM) *GetVMOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get Vm o k response
func (o *GetVMOK) SetPayload(payload *models.VM) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetVMOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetVMBadRequestCode is the HTTP code returned for type GetVMBadRequest
const GetVMBadRequestCode int = 400

/*GetVMBadRequest Invalid ID supplied

swagger:response getVmBadRequest
*/
type GetVMBadRequest struct {
}

// NewGetVMBadRequest creates GetVMBadRequest with default headers values
func NewGetVMBadRequest() *GetVMBadRequest {

	return &GetVMBadRequest{}
}

// WriteResponse to the client
func (o *GetVMBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(400)
}

// GetVMNotFoundCode is the HTTP code returned for type GetVMNotFound
const GetVMNotFoundCode int = 404

/*GetVMNotFound VM not found

swagger:response getVmNotFound
*/
type GetVMNotFound struct {
}

// NewGetVMNotFound creates GetVMNotFound with default headers values
func NewGetVMNotFound() *GetVMNotFound {

	return &GetVMNotFound{}
}

// WriteResponse to the client
func (o *GetVMNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(404)
}