// Code generated by go-swagger; DO NOT EDIT.

package position

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/nntaoli-project/GoEx/bitmex/models"
)

// PositionUpdateLeverageReader is a Reader for the PositionUpdateLeverage structure.
type PositionUpdateLeverageReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PositionUpdateLeverageReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewPositionUpdateLeverageOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 400:
		result := NewPositionUpdateLeverageBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 401:
		result := NewPositionUpdateLeverageUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 404:
		result := NewPositionUpdateLeverageNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewPositionUpdateLeverageOK creates a PositionUpdateLeverageOK with default headers values
func NewPositionUpdateLeverageOK() *PositionUpdateLeverageOK {
	return &PositionUpdateLeverageOK{}
}

/*PositionUpdateLeverageOK handles this case with default header values.

Request was successful
*/
type PositionUpdateLeverageOK struct {
	Payload *models.Position
}

func (o *PositionUpdateLeverageOK) Error() string {
	return fmt.Sprintf("[POST /position/leverage][%d] positionUpdateLeverageOK  %+v", 200, *o.Payload)
}

func (o *PositionUpdateLeverageOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Position)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewPositionUpdateLeverageBadRequest creates a PositionUpdateLeverageBadRequest with default headers values
func NewPositionUpdateLeverageBadRequest() *PositionUpdateLeverageBadRequest {
	return &PositionUpdateLeverageBadRequest{}
}

/*PositionUpdateLeverageBadRequest handles this case with default header values.

Parameter Error
*/
type PositionUpdateLeverageBadRequest struct {
	Payload *models.Error
}

func (o *PositionUpdateLeverageBadRequest) Error() string {
	return fmt.Sprintf("[POST /position/leverage][%d] positionUpdateLeverageBadRequest  %+v", 400, *o.Payload)
}

func (o *PositionUpdateLeverageBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewPositionUpdateLeverageUnauthorized creates a PositionUpdateLeverageUnauthorized with default headers values
func NewPositionUpdateLeverageUnauthorized() *PositionUpdateLeverageUnauthorized {
	return &PositionUpdateLeverageUnauthorized{}
}

/*PositionUpdateLeverageUnauthorized handles this case with default header values.

Unauthorized
*/
type PositionUpdateLeverageUnauthorized struct {
	Payload *models.Error
}

func (o *PositionUpdateLeverageUnauthorized) Error() string {
	return fmt.Sprintf("[POST /position/leverage][%d] positionUpdateLeverageUnauthorized  %+v", 401, o.Payload)
}

func (o *PositionUpdateLeverageUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewPositionUpdateLeverageNotFound creates a PositionUpdateLeverageNotFound with default headers values
func NewPositionUpdateLeverageNotFound() *PositionUpdateLeverageNotFound {
	return &PositionUpdateLeverageNotFound{}
}

/*PositionUpdateLeverageNotFound handles this case with default header values.

Not Found
*/
type PositionUpdateLeverageNotFound struct {
	Payload *models.Error
}

func (o *PositionUpdateLeverageNotFound) Error() string {
	return fmt.Sprintf("[POST /position/leverage][%d] positionUpdateLeverageNotFound  %+v", 404, o.Payload)
}

func (o *PositionUpdateLeverageNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}