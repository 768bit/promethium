// Code generated by go-swagger; DO NOT EDIT.

package vms

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"
)

// New creates a new vms API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) *Client {
	return &Client{transport: transport, formats: formats}
}

/*
Client for vms API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

/*
CreateVM creates a VM instance

Create an instance of VM
*/
func (a *Client) CreateVM(params *CreateVMParams) (*CreateVMOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreateVMParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "createVM",
		Method:             "POST",
		PathPattern:        "/vms",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &CreateVMReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CreateVMOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for createVM: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
CreateVMDisk creates or attach a VM disk

Create or attach a VM Disk
*/
func (a *Client) CreateVMDisk(params *CreateVMDiskParams) (*CreateVMDiskOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreateVMDiskParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "createVMDisk",
		Method:             "POST",
		PathPattern:        "/vms/{vmID}/disks",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &CreateVMDiskReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CreateVMDiskOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*CreateVMDiskDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
CreateVMInterface creates a new VM network itnerface

Create a new VM Network Itnerface
*/
func (a *Client) CreateVMInterface(params *CreateVMInterfaceParams) (*CreateVMInterfaceOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreateVMInterfaceParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "createVMInterface",
		Method:             "POST",
		PathPattern:        "/vms/{vmID}/interfaces",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &CreateVMInterfaceReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CreateVMInterfaceOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*CreateVMInterfaceDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
CreateVMVolume creates or attach a VM volume

Create or attach a VM Volume
*/
func (a *Client) CreateVMVolume(params *CreateVMVolumeParams) (*CreateVMVolumeOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreateVMVolumeParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "createVMVolume",
		Method:             "POST",
		PathPattern:        "/vms/{vmID}/volumes",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &CreateVMVolumeReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CreateVMVolumeOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*CreateVMVolumeDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
DeleteVM destroys a VM instance

Destroy an isntance of VM
*/
func (a *Client) DeleteVM(params *DeleteVMParams) (*DeleteVMOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeleteVMParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "deleteVM",
		Method:             "DELETE",
		PathPattern:        "/vms/{vmID}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &DeleteVMReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*DeleteVMOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for deleteVM: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
DeleteVMDrive returns a VM instance

Returns an isntance of VM
*/
func (a *Client) DeleteVMDrive(params *DeleteVMDriveParams) (*DeleteVMDriveOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeleteVMDriveParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "deleteVMDrive",
		Method:             "DELETE",
		PathPattern:        "/vms/{vmID}/disks/{diskID}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &DeleteVMDriveReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*DeleteVMDriveOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*DeleteVMDriveDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
DeleteVMInterface destroys a VM network interface

Destroy a VM Network Interface
*/
func (a *Client) DeleteVMInterface(params *DeleteVMInterfaceParams) (*DeleteVMInterfaceOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeleteVMInterfaceParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "deleteVMInterface",
		Method:             "DELETE",
		PathPattern:        "/vms/{vmID}/interfaces/{interfaceID}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &DeleteVMInterfaceReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*DeleteVMInterfaceOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*DeleteVMInterfaceDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
DeleteVMVolume returns a VM instance

Returns an isntance of VM
*/
func (a *Client) DeleteVMVolume(params *DeleteVMVolumeParams) (*DeleteVMVolumeOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeleteVMVolumeParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "deleteVMVolume",
		Method:             "DELETE",
		PathPattern:        "/vms/{vmID}/volumes/{volumeID}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &DeleteVMVolumeReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*DeleteVMVolumeOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*DeleteVMVolumeDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
GetVM returns a VM instance

Returns an isntance of VM
*/
func (a *Client) GetVM(params *GetVMParams) (*GetVMOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVMParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVM",
		Method:             "GET",
		PathPattern:        "/vms/{vmID}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetVMReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVMOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getVM: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetVMConsole gets a console for a VM instance

Get a console for a VM instance
*/
func (a *Client) GetVMConsole(params *GetVMConsoleParams) (*GetVMConsoleOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVMConsoleParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVMConsole",
		Method:             "GET",
		PathPattern:        "/vms/{vmID}/console",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetVMConsoleReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVMConsoleOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getVMConsole: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetVMDisk returns a VM disk

Returns an isntance of VM Disk
*/
func (a *Client) GetVMDisk(params *GetVMDiskParams) (*GetVMDiskOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVMDiskParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVMDisk",
		Method:             "GET",
		PathPattern:        "/vms/{vmID}/disks/{diskID}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetVMDiskReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVMDiskOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetVMDiskDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
GetVMDiskList gets a list of VM attached disks

Returns a list of VM Attached Disks
*/
func (a *Client) GetVMDiskList(params *GetVMDiskListParams) (*GetVMDiskListOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVMDiskListParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVMDiskList",
		Method:             "GET",
		PathPattern:        "/vms/{vmID}/disks",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetVMDiskListReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVMDiskListOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetVMDiskListDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
GetVMInterace returns a VM network interface

Returns an instance of VM Network Interface
*/
func (a *Client) GetVMInterace(params *GetVMInteraceParams) (*GetVMInteraceOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVMInteraceParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVMInterace",
		Method:             "GET",
		PathPattern:        "/vms/{vmID}/interfaces/{interfaceID}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetVMInteraceReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVMInteraceOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetVMInteraceDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
GetVMInterfaceList gets a list of VM network interfaces

Returns a list of VM Network Itnerfaces
*/
func (a *Client) GetVMInterfaceList(params *GetVMInterfaceListParams) (*GetVMInterfaceListOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVMInterfaceListParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVMInterfaceList",
		Method:             "GET",
		PathPattern:        "/vms/{vmID}/interfaces",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetVMInterfaceListReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVMInterfaceListOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetVMInterfaceListDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
GetVMList gets a list of v ms

Returns a list of VMs
*/
func (a *Client) GetVMList(params *GetVMListParams) (*GetVMListOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVMListParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVMList",
		Method:             "GET",
		PathPattern:        "/vms",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetVMListReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVMListOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetVMListDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
GetVMVolume returns a VM instance

Returns an isntance of VM
*/
func (a *Client) GetVMVolume(params *GetVMVolumeParams) (*GetVMVolumeOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVMVolumeParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVMVolume",
		Method:             "GET",
		PathPattern:        "/vms/{vmID}/volumes/{volumeID}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetVMVolumeReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVMVolumeOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetVMVolumeDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
GetVMVolumeList gets a list of VM volumes

Returns a list of VM Volumes
*/
func (a *Client) GetVMVolumeList(params *GetVMVolumeListParams) (*GetVMVolumeListOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVMVolumeListParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVMVolumeList",
		Method:             "GET",
		PathPattern:        "/vms/{vmID}/volumes",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetVMVolumeListReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVMVolumeListOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetVMVolumeListDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
ResetVM resets a VM instance

Forcefully Reset an instance of VM
*/
func (a *Client) ResetVM(params *ResetVMParams) (*ResetVMOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewResetVMParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "resetVM",
		Method:             "GET",
		PathPattern:        "/vms/{vmID}/reset",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &ResetVMReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ResetVMOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for resetVM: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
RestartVM restarts a VM instance

Gracefully Restart an instance of VM
*/
func (a *Client) RestartVM(params *RestartVMParams) (*RestartVMOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewRestartVMParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "restartVM",
		Method:             "GET",
		PathPattern:        "/vms/{vmID}/restart",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &RestartVMReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*RestartVMOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for restartVM: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
ShutdownVM shutdowns a VM instance

Gracefully Shutdown an instance of VM
*/
func (a *Client) ShutdownVM(params *ShutdownVMParams) (*ShutdownVMOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewShutdownVMParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "shutdownVM",
		Method:             "GET",
		PathPattern:        "/vms/{vmID}/shutdown",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &ShutdownVMReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ShutdownVMOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for shutdownVM: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
StartVM starts a VM instance

Starts an isntance of VM
*/
func (a *Client) StartVM(params *StartVMParams) (*StartVMOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewStartVMParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "startVM",
		Method:             "GET",
		PathPattern:        "/vms/{vmID}/start",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &StartVMReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*StartVMOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for startVM: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
StopVM stops a VM instance

Stops an isntance of VM
*/
func (a *Client) StopVM(params *StopVMParams) (*StopVMOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewStopVMParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "stopVM",
		Method:             "GET",
		PathPattern:        "/vms/{vmID}/stop",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &StopVMReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*StopVMOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for stopVM: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
UpdateVM updates a VM instance

Update an instance of VM
*/
func (a *Client) UpdateVM(params *UpdateVMParams) (*UpdateVMOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewUpdateVMParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "updateVM",
		Method:             "PUT",
		PathPattern:        "/vms/{vmID}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &UpdateVMReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*UpdateVMOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for updateVM: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
UpdateVMDisk updates a VM interface instance

Update an instance of VM interface
*/
func (a *Client) UpdateVMDisk(params *UpdateVMDiskParams) (*UpdateVMDiskOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewUpdateVMDiskParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "updateVMDisk",
		Method:             "PUT",
		PathPattern:        "/vms/{vmID}/disks/{diskID}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &UpdateVMDiskReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*UpdateVMDiskOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*UpdateVMDiskDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
UpdateVMInterface updates a VM network interface instance

Update an instance of VM Network Interface
*/
func (a *Client) UpdateVMInterface(params *UpdateVMInterfaceParams) (*UpdateVMInterfaceOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewUpdateVMInterfaceParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "updateVMInterface",
		Method:             "PUT",
		PathPattern:        "/vms/{vmID}/interfaces/{interfaceID}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &UpdateVMInterfaceReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*UpdateVMInterfaceOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*UpdateVMInterfaceDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
UpdateVMVolume updates a VM interface instance

Update an instance of VM interface
*/
func (a *Client) UpdateVMVolume(params *UpdateVMVolumeParams) (*UpdateVMVolumeOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewUpdateVMVolumeParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "updateVMVolume",
		Method:             "PUT",
		PathPattern:        "/vms/{vmID}/volumes/{volumeID}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &UpdateVMVolumeReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*UpdateVMVolumeOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*UpdateVMVolumeDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
