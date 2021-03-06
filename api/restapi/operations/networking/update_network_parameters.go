// Code generated by go-swagger; DO NOT EDIT.

package networking

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/768bit/promethium/api/models"
)

// NewUpdateNetworkParams creates a new UpdateNetworkParams object
// no default values defined in spec.
func NewUpdateNetworkParams() UpdateNetworkParams {

	return UpdateNetworkParams{}
}

// UpdateNetworkParams contains all the bound params for the update network operation
// typically these are obtained from a http.Request
//
// swagger:parameters updateNetwork
type UpdateNetworkParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*Create new VM instance
	  Required: true
	  In: body
	*/
	NetConfig models.UpdateNetwork
	/*ID of VM to return
	  Required: true
	  In: path
	*/
	NetworkID string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewUpdateNetworkParams() beforehand.
func (o *UpdateNetworkParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.UpdateNetwork
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("netConfig", "body"))
			} else {
				res = append(res, errors.NewParseError("netConfig", "body", "", err))
			}
		} else {
			// no validation on generic interface
			o.NetConfig = body
		}
	} else {
		res = append(res, errors.Required("netConfig", "body"))
	}
	rNetworkID, rhkNetworkID, _ := route.Params.GetOK("networkID")
	if err := o.bindNetworkID(rNetworkID, rhkNetworkID, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindNetworkID binds and validates parameter NetworkID from path.
func (o *UpdateNetworkParams) bindNetworkID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	o.NetworkID = raw

	return nil
}
