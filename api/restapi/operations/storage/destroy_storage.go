// Code generated by go-swagger; DO NOT EDIT.

package storage

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// DestroyStorageHandlerFunc turns a function with the right signature into a destroy storage handler
type DestroyStorageHandlerFunc func(DestroyStorageParams) middleware.Responder

// Handle executing the request and returning a response
func (fn DestroyStorageHandlerFunc) Handle(params DestroyStorageParams) middleware.Responder {
	return fn(params)
}

// DestroyStorageHandler interface for that can handle valid destroy storage params
type DestroyStorageHandler interface {
	Handle(DestroyStorageParams) middleware.Responder
}

// NewDestroyStorage creates a new http.Handler for the destroy storage operation
func NewDestroyStorage(ctx *middleware.Context, handler DestroyStorageHandler) *DestroyStorage {
	return &DestroyStorage{Context: ctx, Handler: handler}
}

/*DestroyStorage swagger:route DELETE /storage/{storageID} storage destroyStorage

Remove storage item

*/
type DestroyStorage struct {
	Context *middleware.Context
	Handler DestroyStorageHandler
}

func (o *DestroyStorage) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewDestroyStorageParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}