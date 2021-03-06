// Code generated by go-swagger; DO NOT EDIT.

package images

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/768bit/promethium/api/models"
)

// PushImageOKCode is the HTTP code returned for type PushImageOK
const PushImageOKCode int = 200

/*PushImageOK successful operation

swagger:response pushImageOK
*/
type PushImageOK struct {

	/*
	  In: Body
	*/
	Payload *models.VM `json:"body,omitempty"`
}

// NewPushImageOK creates PushImageOK with default headers values
func NewPushImageOK() *PushImageOK {

	return &PushImageOK{}
}

// WithPayload adds the payload to the push image o k response
func (o *PushImageOK) WithPayload(payload *models.VM) *PushImageOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the push image o k response
func (o *PushImageOK) SetPayload(payload *models.VM) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PushImageOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*PushImageDefault unexpected error

swagger:response pushImageDefault
*/
type PushImageDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPushImageDefault creates PushImageDefault with default headers values
func NewPushImageDefault(code int) *PushImageDefault {
	if code <= 0 {
		code = 500
	}

	return &PushImageDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the push image default response
func (o *PushImageDefault) WithStatusCode(code int) *PushImageDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the push image default response
func (o *PushImageDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the push image default response
func (o *PushImageDefault) WithPayload(payload *models.Error) *PushImageDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the push image default response
func (o *PushImageDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PushImageDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
