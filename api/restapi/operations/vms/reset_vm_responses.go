// Code generated by go-swagger; DO NOT EDIT.

package vms

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/768bit/promethium/api/models"
)

// ResetVMOKCode is the HTTP code returned for type ResetVMOK
const ResetVMOKCode int = 200

/*ResetVMOK successful operation

swagger:response resetVmOK
*/
type ResetVMOK struct {

	/*
	  In: Body
	*/
	Payload *models.VM `json:"body,omitempty"`
}

// NewResetVMOK creates ResetVMOK with default headers values
func NewResetVMOK() *ResetVMOK {

	return &ResetVMOK{}
}

// WithPayload adds the payload to the reset Vm o k response
func (o *ResetVMOK) WithPayload(payload *models.VM) *ResetVMOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the reset Vm o k response
func (o *ResetVMOK) SetPayload(payload *models.VM) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ResetVMOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// ResetVMBadRequestCode is the HTTP code returned for type ResetVMBadRequest
const ResetVMBadRequestCode int = 400

/*ResetVMBadRequest Invalid ID supplied

swagger:response resetVmBadRequest
*/
type ResetVMBadRequest struct {
}

// NewResetVMBadRequest creates ResetVMBadRequest with default headers values
func NewResetVMBadRequest() *ResetVMBadRequest {

	return &ResetVMBadRequest{}
}

// WriteResponse to the client
func (o *ResetVMBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(400)
}

// ResetVMNotFoundCode is the HTTP code returned for type ResetVMNotFound
const ResetVMNotFoundCode int = 404

/*ResetVMNotFound VM not found

swagger:response resetVmNotFound
*/
type ResetVMNotFound struct {
}

// NewResetVMNotFound creates ResetVMNotFound with default headers values
func NewResetVMNotFound() *ResetVMNotFound {

	return &ResetVMNotFound{}
}

// WriteResponse to the client
func (o *ResetVMNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(404)
}