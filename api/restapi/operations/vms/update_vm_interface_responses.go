// Code generated by go-swagger; DO NOT EDIT.

package vms

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/768bit/promethium/api/models"
)

// UpdateVMInterfaceOKCode is the HTTP code returned for type UpdateVMInterfaceOK
const UpdateVMInterfaceOKCode int = 200

/*UpdateVMInterfaceOK successful operation

swagger:response updateVmInterfaceOK
*/
type UpdateVMInterfaceOK struct {

	/*
	  In: Body
	*/
	Payload *models.MetaDataNetworkInterfaceConfig `json:"body,omitempty"`
}

// NewUpdateVMInterfaceOK creates UpdateVMInterfaceOK with default headers values
func NewUpdateVMInterfaceOK() *UpdateVMInterfaceOK {

	return &UpdateVMInterfaceOK{}
}

// WithPayload adds the payload to the update Vm interface o k response
func (o *UpdateVMInterfaceOK) WithPayload(payload *models.MetaDataNetworkInterfaceConfig) *UpdateVMInterfaceOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update Vm interface o k response
func (o *UpdateVMInterfaceOK) SetPayload(payload *models.MetaDataNetworkInterfaceConfig) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateVMInterfaceOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UpdateVMInterfaceBadRequestCode is the HTTP code returned for type UpdateVMInterfaceBadRequest
const UpdateVMInterfaceBadRequestCode int = 400

/*UpdateVMInterfaceBadRequest Invalid ID supplied

swagger:response updateVmInterfaceBadRequest
*/
type UpdateVMInterfaceBadRequest struct {
}

// NewUpdateVMInterfaceBadRequest creates UpdateVMInterfaceBadRequest with default headers values
func NewUpdateVMInterfaceBadRequest() *UpdateVMInterfaceBadRequest {

	return &UpdateVMInterfaceBadRequest{}
}

// WriteResponse to the client
func (o *UpdateVMInterfaceBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(400)
}

// UpdateVMInterfaceNotFoundCode is the HTTP code returned for type UpdateVMInterfaceNotFound
const UpdateVMInterfaceNotFoundCode int = 404

/*UpdateVMInterfaceNotFound VM Interface not found

swagger:response updateVmInterfaceNotFound
*/
type UpdateVMInterfaceNotFound struct {
}

// NewUpdateVMInterfaceNotFound creates UpdateVMInterfaceNotFound with default headers values
func NewUpdateVMInterfaceNotFound() *UpdateVMInterfaceNotFound {

	return &UpdateVMInterfaceNotFound{}
}

// WriteResponse to the client
func (o *UpdateVMInterfaceNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(404)
}

/*UpdateVMInterfaceDefault unexpected error

swagger:response updateVmInterfaceDefault
*/
type UpdateVMInterfaceDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewUpdateVMInterfaceDefault creates UpdateVMInterfaceDefault with default headers values
func NewUpdateVMInterfaceDefault(code int) *UpdateVMInterfaceDefault {
	if code <= 0 {
		code = 500
	}

	return &UpdateVMInterfaceDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the update VM interface default response
func (o *UpdateVMInterfaceDefault) WithStatusCode(code int) *UpdateVMInterfaceDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the update VM interface default response
func (o *UpdateVMInterfaceDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the update VM interface default response
func (o *UpdateVMInterfaceDefault) WithPayload(payload *models.Error) *UpdateVMInterfaceDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update VM interface default response
func (o *UpdateVMInterfaceDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateVMInterfaceDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
