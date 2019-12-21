// Code generated by go-swagger; DO NOT EDIT.

package vms

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// DeleteVMInterfaceHandlerFunc turns a function with the right signature into a delete VM interface handler
type DeleteVMInterfaceHandlerFunc func(DeleteVMInterfaceParams) middleware.Responder

// Handle executing the request and returning a response
func (fn DeleteVMInterfaceHandlerFunc) Handle(params DeleteVMInterfaceParams) middleware.Responder {
	return fn(params)
}

// DeleteVMInterfaceHandler interface for that can handle valid delete VM interface params
type DeleteVMInterfaceHandler interface {
	Handle(DeleteVMInterfaceParams) middleware.Responder
}

// NewDeleteVMInterface creates a new http.Handler for the delete VM interface operation
func NewDeleteVMInterface(ctx *middleware.Context, handler DeleteVMInterfaceHandler) *DeleteVMInterface {
	return &DeleteVMInterface{Context: ctx, Handler: handler}
}

/*DeleteVMInterface swagger:route DELETE /vms/{vmID}/interfaces/{interfaceID} vms deleteVmInterface

Destroy a VM Network Interface

Destroy a VM Network Interface

*/
type DeleteVMInterface struct {
	Context *middleware.Context
	Handler DeleteVMInterfaceHandler
}

func (o *DeleteVMInterface) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewDeleteVMInterfaceParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}