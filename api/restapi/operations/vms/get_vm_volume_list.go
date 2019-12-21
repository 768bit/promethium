// Code generated by go-swagger; DO NOT EDIT.

package vms

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// GetVMVolumeListHandlerFunc turns a function with the right signature into a get VM volume list handler
type GetVMVolumeListHandlerFunc func(GetVMVolumeListParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetVMVolumeListHandlerFunc) Handle(params GetVMVolumeListParams) middleware.Responder {
	return fn(params)
}

// GetVMVolumeListHandler interface for that can handle valid get VM volume list params
type GetVMVolumeListHandler interface {
	Handle(GetVMVolumeListParams) middleware.Responder
}

// NewGetVMVolumeList creates a new http.Handler for the get VM volume list operation
func NewGetVMVolumeList(ctx *middleware.Context, handler GetVMVolumeListHandler) *GetVMVolumeList {
	return &GetVMVolumeList{Context: ctx, Handler: handler}
}

/*GetVMVolumeList swagger:route GET /vms/{vmID}/volumes vms getVmVolumeList

Get a list of VM Volumes

Returns a list of VM Volumes

*/
type GetVMVolumeList struct {
	Context *middleware.Context
	Handler GetVMVolumeListHandler
}

func (o *GetVMVolumeList) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetVMVolumeListParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}