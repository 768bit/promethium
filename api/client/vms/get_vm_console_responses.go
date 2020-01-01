// Code generated by go-swagger; DO NOT EDIT.

package vms

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/768bit/promethium/api/models"
)

// GetVMConsoleReader is a Reader for the GetVMConsole structure.
type GetVMConsoleReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetVMConsoleReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetVMConsoleOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetVMConsoleBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetVMConsoleNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetVMConsoleOK creates a GetVMConsoleOK with default headers values
func NewGetVMConsoleOK() *GetVMConsoleOK {
	return &GetVMConsoleOK{}
}

/*GetVMConsoleOK handles this case with default header values.

successful operation
*/
type GetVMConsoleOK struct {
	Payload *models.VM
}

func (o *GetVMConsoleOK) Error() string {
	return fmt.Sprintf("[GET /vms/{vmID}/console][%d] getVmConsoleOK  %+v", 200, o.Payload)
}

func (o *GetVMConsoleOK) GetPayload() *models.VM {
	return o.Payload
}

func (o *GetVMConsoleOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.VM)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetVMConsoleBadRequest creates a GetVMConsoleBadRequest with default headers values
func NewGetVMConsoleBadRequest() *GetVMConsoleBadRequest {
	return &GetVMConsoleBadRequest{}
}

/*GetVMConsoleBadRequest handles this case with default header values.

Invalid ID supplied
*/
type GetVMConsoleBadRequest struct {
}

func (o *GetVMConsoleBadRequest) Error() string {
	return fmt.Sprintf("[GET /vms/{vmID}/console][%d] getVmConsoleBadRequest ", 400)
}

func (o *GetVMConsoleBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewGetVMConsoleNotFound creates a GetVMConsoleNotFound with default headers values
func NewGetVMConsoleNotFound() *GetVMConsoleNotFound {
	return &GetVMConsoleNotFound{}
}

/*GetVMConsoleNotFound handles this case with default header values.

VM not found
*/
type GetVMConsoleNotFound struct {
}

func (o *GetVMConsoleNotFound) Error() string {
	return fmt.Sprintf("[GET /vms/{vmID}/console][%d] getVmConsoleNotFound ", 404)
}

func (o *GetVMConsoleNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
