package uber

import (
	"bytes"
	"fmt"
)

const (
	Version         = "v1"
	RequestEndpoint = "requests"
	// the next two use `AUTH_EDPOINT`
	AccessCodeEndpoint  = "authorize"
	AccessTokenEndpoint = "token"
	State = "go-uber"
	Port  = ":7635"
)


var (
	AuthHost    = "https://login.uber.com/oauth"
	UberSandboxAPIHost = fmt.Sprintf("https://sandbox-api.uber.com/%s", Version)
	UberAPIHost = fmt.Sprintf("https://sandbox-api.uber.com/%s", Version)
)

//
// structs representing the necessary data for generating requests to the various
// endpoints
//

type authReq struct {
	auth
	responseType string `query:"response_type,required"`
	scope        string `query:"scope"`
	state        string `query:"state"`
}

type accReq struct {
	auth
	clientSecret string `query:"client_secret,required"`
	grantType    string `query:"grant_type,required"`
	code         string `query:"code,required"`
}

type requestReq struct {
	productID           string  `query:"product_id,required"`
	startLatitude       float64 `query:"start_latitude,required"`
	startLongitude      float64 `query:"start_longitude,required"`
	endLatitude         float64 `query:"end_latitude,required"`
	endLongitude        float64 `query:"end_longitude,required"`
	surgeConfirmationID string  `query:"surge_confirmation_id"`
}

type requestResp struct {
	Request
}

type requestMapResp struct {
	RequestID string `json:"request_id"`
	HRef      string `json:"href"`
}

// Request contains the information relating to a request for an Uber done on behalf of a
// user.
type Request struct {
	RequestID       string `json:"request_id"`
	Status          string `json:"status"`
	Vehicle         `json:"vehicle"`
	Driver          `json:"driver"`
	Location        `json:"location"`
	ETA             int     `json:"eta"`
	SurgeMultiplier float64 `json:"surge_multiplier"`
}

// Vehicle represents the car in a response to requesting a ride.
type Vehicle struct {
	Make         string `json:"make"`
	Model        string `json:"model"`
	LicensePlate string `json:"license_plate"`
	PictureURL   string `json:"picture_url"`
}

// Driver represents an Uber driver.
type Driver struct {
	PhoneNumber string `json:"phone_number"`
	Rating      int    `json:"rating"`
	PictureURL  string `json:"picture_url"`
	Name        string `json:"name"`
}

// Location contains a human-readable address as well as the exact coordinates of a location.
type Location struct {
	Address string `json:"address,omitempty"`
	Latitude float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// uberError implements the error interface (by defining an `Error() string` method).
// This datatype is returned from the Uber API with non-2xx responses.
type uberError struct {
	Message string `json:"message"`
	Code string `json:"code"`
	Fields map[string]string `json:"fields,omitempty"`
}

// Error implements the `error` interface for `uberError`.
func (err uberError) Error() string {
	var uberErrBuff bytes.Buffer // because O(1) runtime, bitches
	uberErrBuff.WriteString(fmt.Sprintf("Uber API: %s", err.Message))

	// prints code if exists
	if err.Code != "" {
		uberErrBuff.WriteString(fmt.Sprintf("\nCode: %s", err.Code))
	}

	// prints erroneous fields
	if err.Fields != nil {
		uberErrBuff.WriteString("\nFields:")
		for k, v := range err.Fields {
			uberErrBuff.WriteString(fmt.Sprintf("\n\t%s: %v", k, v))
		}
	}

	return uberErrBuff.String()
}
