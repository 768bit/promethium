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

// GetVMReader is a Reader for the GetVM structure.
type GetVMReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetVMReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetVMOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetVMBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetVMNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetVMOK creates a GetVMOK with default headers values
func NewGetVMOK() *GetVMOK {
	return &GetVMOK{}
}

/*GetVMOK handles this case with default header values.

successful operation
*/
type GetVMOK struct {
	Payload *models.VM
}

func (o *GetVMOK) Error() string {
	return fmt.Sprintf("[GET /vms/{vmID}][%d] getVmOK  %+v", 200, o.Payload)
}

func (o *GetVMOK) GetPayload() *models.VM {
	return o.Payload
}

func (o *GetVMOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.VM)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetVMBadRequest creates a GetVMBadRequest with default headers values
func NewGetVMBadRequest() *GetVMBadRequest {
	return &GetVMBadRequest{}
}

/*GetVMBadRequest handles this case with default header values.

Invalid ID supplied
*/
type GetVMBadRequest struct {
}

func (o *GetVMBadRequest) Error() string {
	return fmt.Sprintf("[GET /vms/{vmID}][%d] getVmBadRequest ", 400)
}

func (o *GetVMBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewGetVMNotFound creates a GetVMNotFound with default headers values
func NewGetVMNotFound() *GetVMNotFound {
	return &GetVMNotFound{}
}

/*GetVMNotFound handles this case with default header values.

VM not found
*/
type GetVMNotFound struct {
}

func (o *GetVMNotFound) Error() string {
	return fmt.Sprintf("[GET /vms/{vmID}][%d] getVmNotFound ", 404)
}

func (o *GetVMNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
