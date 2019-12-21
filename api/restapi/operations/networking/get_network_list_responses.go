// Code generated by go-swagger; DO NOT EDIT.

package networking

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/768bit/promethium/api/models"
)

// GetNetworkListOKCode is the HTTP code returned for type GetNetworkListOK
const GetNetworkListOKCode int = 200

/*GetNetworkListOK OK

swagger:response getNetworkListOK
*/
type GetNetworkListOK struct {

	/*
	  In: Body
	*/
	Payload []models.Network `json:"body,omitempty"`
}

// NewGetNetworkListOK creates GetNetworkListOK with default headers values
func NewGetNetworkListOK() *GetNetworkListOK {

	return &GetNetworkListOK{}
}

// WithPayload adds the payload to the get network list o k response
func (o *GetNetworkListOK) WithPayload(payload []models.Network) *GetNetworkListOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get network list o k response
func (o *GetNetworkListOK) SetPayload(payload []models.Network) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetNetworkListOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty array
		payload = make([]models.Network, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}