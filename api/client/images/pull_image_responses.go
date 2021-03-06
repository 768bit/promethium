// Code generated by go-swagger; DO NOT EDIT.

package images

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/768bit/promethium/api/models"
)

// PullImageReader is a Reader for the PullImage structure.
type PullImageReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PullImageReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewPullImageOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewPullImageDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewPullImageOK creates a PullImageOK with default headers values
func NewPullImageOK() *PullImageOK {
	return &PullImageOK{}
}

/*PullImageOK handles this case with default header values.

successful operation
*/
type PullImageOK struct {
	Payload *models.VM
}

func (o *PullImageOK) Error() string {
	return fmt.Sprintf("[POST /images/pull][%d] pullImageOK  %+v", 200, o.Payload)
}

func (o *PullImageOK) GetPayload() *models.VM {
	return o.Payload
}

func (o *PullImageOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.VM)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewPullImageDefault creates a PullImageDefault with default headers values
func NewPullImageDefault(code int) *PullImageDefault {
	return &PullImageDefault{
		_statusCode: code,
	}
}

/*PullImageDefault handles this case with default header values.

unexpected error
*/
type PullImageDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the pull image default response
func (o *PullImageDefault) Code() int {
	return o._statusCode
}

func (o *PullImageDefault) Error() string {
	return fmt.Sprintf("[POST /images/pull][%d] pullImage default  %+v", o._statusCode, o.Payload)
}

func (o *PullImageDefault) GetPayload() *models.Error {
	return o.Payload
}

func (o *PullImageDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
