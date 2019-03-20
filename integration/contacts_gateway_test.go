// +build integration

package integration

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/bitly/go-simplejson"
	"github.com/infobloxopen/atlas-contacts-app/cmd"
	"github.com/infobloxopen/atlas-contacts-app/pkg/pb"
)

// TestCreateContact_REST uses the REST gateway to create a new contact and
// ensure the JSON response matches what is expected
// 1. Create a contact entry with a POST request
// 2. Unmarshal the JSON into a simplejson struct
// 3. Ensure the JSON fields match what is expected
func TestCreateContact_REST(t *testing.T) {
	dbTest.Reset(t)
	contact := pb.Contact{
		FirstName:    "Steven",
		MiddleName:   "James",
		LastName:     "McKubernetes",
		PrimaryEmail: "test@test.com",
		Notes:        "set sail at sunrise",
	}
	resCreate, err := MakeRequestWithDefaults(
		http.MethodPost,
		"http://localhost:8080/v1/contacts",
		contact,
	)
	if err != nil {
		t.Fatalf("unable to create contact: %v", err)
	}
	ValidateResponseCode(t, resCreate, http.StatusOK)
	createJSON, err := simplejson.NewFromReader(resCreate.Body)
	if err != nil {
		t.Fatalf("unable to marshal json response: %v", err)
	}
	var tests = []struct {
		name   string
		json   *simplejson.Json
		expect string
	}{
		{
			name:   "contact first name",
			json:   createJSON.GetPath("result", "first_name"),
			expect: `"Steven"`,
		},
		{
			name:   "contact middle name",
			json:   createJSON.GetPath("result", "middle_name"),
			expect: `"James"`,
		},
		{
			name:   "contact last",
			json:   createJSON.GetPath("result", "last_name"),
			expect: `"McKubernetes"`,
		},
		{
			name:   "contact notes",
			json:   createJSON.GetPath("result", "notes"),
			expect: `"set sail at sunrise"`,
		},
		{
			name:   "contact primary email",
			json:   createJSON.GetPath("result", "primary_email"),
			expect: `"test@test.com"`,
		},
		{
			name:   "contact email list",
			json:   createJSON.GetPath("result", "emails"),
			expect: `[{"address":"test@test.com","id":"1"}]`,
		},
		{
			name:   "success response",
			json:   createJSON.GetPath("success"),
			expect: `{"code":"OK","status":200}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ValidateJSONSchema(t, test.json, test.expect)
		})
	}
}

// TestReadContact_REST uses the REST gateway to create a new contact and
// then read that contact from the application
// 1. Create a contact entry with a POST request
// 2. Get the contact from the applicaiton
// 2. Unmarshal the JSON into a simplejson struct
// 3. Ensure the JSON fields match what is expected
func TestReadContact_REST(t *testing.T) {
	dbTest.Reset(t)
	contact := pb.Contact{
		FirstName:    "Wilfred",
		MiddleName:   "Wallace",
		LastName:     "O'Docker",
		PrimaryEmail: "test@test.com",
		Notes:        "build the containers at dusk",
	}
	resCreate, err := MakeRequestWithDefaults(
		http.MethodPost, "http://localhost:8080/v1/contacts",
		contact,
	)
	if err != nil {
		t.Fatalf("unable to create contact: %v", err)
	}
	createJSON, err := simplejson.NewFromReader(resCreate.Body)
	if err != nil {
		t.Fatalf("unable to unmarshal create contact response body: %v", err)
	}
	id, err := createJSON.GetPath("result", "id").String()
	if err != nil {
		t.Fatalf("unable to get contact id from response json: %v", err)
	}
	id = strings.TrimPrefix(id, fmt.Sprintf("%s/%s/", cmd.ApplicationID, "contact"))
	resRead, err := MakeRequestWithDefaults(
		http.MethodGet, fmt.Sprintf("http://localhost:8080/v1/contacts/%s", id),
		nil,
	)
	if err != nil {
		t.Fatalf("unable to get contact: %v", err)
	}
	ValidateResponseCode(t, resRead, http.StatusOK)
	readJSON, err := simplejson.NewFromReader(resRead.Body)
	var tests = []struct {
		name   string
		json   *simplejson.Json
		expect string
	}{
		{
			name:   "contact first name",
			json:   readJSON.GetPath("result", "first_name"),
			expect: `"Wilfred"`,
		},
		{
			name:   "contact middle name",
			json:   readJSON.GetPath("result", "middle_name"),
			expect: `"Wallace"`,
		},
		{
			name:   "contact last",
			json:   readJSON.GetPath("result", "last_name"),
			expect: `"O'Docker"`,
		},
		{
			name:   "contact notes",
			json:   readJSON.GetPath("result", "notes"),
			expect: `"build the containers at dusk"`,
		},
		{
			name:   "contact primary email",
			json:   readJSON.GetPath("result", "primary_email"),
			expect: `"test@test.com"`,
		},
		{
			name:   "contact email list",
			json:   readJSON.GetPath("result", "emails"),
			expect: `[{"address":"test@test.com","id":"1"}]`,
		},
		{
			name:   "success response",
			json:   readJSON.GetPath("success"),
			expect: `{"code":"OK","status":200}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ValidateJSONSchema(t, test.json, test.expect)
		})
	}
}

// TestInvalidEmail_REST attempts to create a new contact that has an invalid
// email address.
// 1. Use the REST API to create a new contact with an invalid email
// 2. Ensure the HTTP response status code is 400
// 3. Ensure the HTTP response has a detailed error message
func TestInvalidEmail_REST(t *testing.T) {
	dbTest.Reset(t)
	contact := pb.Contact{
		PrimaryEmail: "invalid-email-address",
	}
	resDelete, err := MakeRequestWithDefaults(
		http.MethodPost, "http://localhost:8080/v1/contacts",
		contact,
	)
	if err != nil {
		t.Fatalf("unable to create contact %v", err)
	}
	ValidateResponseCode(t, resDelete, http.StatusBadRequest)
	deleteJSON, err := simplejson.NewFromReader(resDelete.Body)
	if err != nil {
		t.Fatalf("unable to unmarshal response json: %v", err)
	}
	var tests = []struct {
		name   string
		json   *simplejson.Json
		expect string
	}{
		{
			name:   "check response code",
			json:   deleteJSON.GetPath("error", "code"),
			expect: `"INVALID_ARGUMENT"`,
		},
		{
			name:   "check http status",
			json:   deleteJSON.GetPath("error", "status"),
			expect: `400`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ValidateJSONSchema(t, test.json, test.expect)
		})
	}
}

// TestDeleteContact_REST uses the REST gateway to create a new contact and
// ensure it can get deleted correctly
// 1. Create a contact entry with a POST request
// 2. Delete the contact
// 3. Ensure the DELETE response matches schema is expected
func TestDeleteContact_REST(t *testing.T) {
	dbTest.Reset(t)
	contact := pb.Contact{
		PrimaryEmail: "test@test.com",
	}
	resCreate, err := MakeRequestWithDefaults(
		http.MethodPost,
		"http://localhost:8080/v1/contacts",
		contact,
	)
	if err != nil {
		t.Fatalf("unable to create contact: %v", err)
	}
	createJSON, err := simplejson.NewFromReader(resCreate.Body)
	if err != nil {
		t.Fatalf("unable to unmarshal create contact response body: %v", err)
	}
	id, err := createJSON.GetPath("result", "id").String()
	if err != nil {
		t.Fatalf("unable to get contact id from response json: %v", err)
	}
	id = strings.TrimPrefix(id, fmt.Sprintf("%s/%s/", cmd.ApplicationID, "contact"))
	resDelete, err := MakeRequestWithDefaults(
		http.MethodDelete,
		fmt.Sprintf("http://localhost:8080/v1/contacts/%s", id),
		nil,
	)
	if err != nil {
		t.Fatalf("unable to delete contact: %v", err)
	}
	ValidateResponseCode(t, resDelete, http.StatusOK)
	deleteJSON, err := simplejson.NewFromReader(resDelete.Body)
	if err != nil {
		t.Fatalf("unable to marshal json response: %v", err)
	}
	t.Run("success response", func(t *testing.T) {
		ValidateJSONSchema(t, deleteJSON, `{"success":{"code":"OK","status":200}}`)
	})
}

// ValidateResponseCode checks the http status of a given request and will
// fail the current test if it doesn't match the expected status code
func ValidateResponseCode(t *testing.T, res *http.Response, expected int) {
	if expected != res.StatusCode {
		t.Errorf("validation error: unexpected http response status: have %d; want %d",
			res.StatusCode, expected,
		)
	}
}

// ValidateJSONSchema ensures a given json field matches an expcted json
// string
func ValidateJSONSchema(t *testing.T, json *simplejson.Json, expected string) {
	if json == nil {
		t.Fatalf("validation error: json schema for is nil")
	}
	encoded, err := json.Encode()
	if err != nil {
		t.Fatalf("validation error: unable to encode expected json: %v", err)
	}
	if actual := string(encoded); actual != expected {
		t.Errorf("actual json schema does not match expected schema: have %s; want %v",
			actual, expected,
		)
	}
}
