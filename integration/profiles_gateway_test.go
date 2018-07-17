// +build integration

package integration

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/infobloxopen/atlas-contacts-app/cmd"
	"github.com/infobloxopen/atlas-contacts-app/pkg/pb"
)

// TestCreateProfile_gateway uses the REST gateway to create a new profile and
// ensure the JSON response matches is expected
// 1. Create a profile entry with a POST request to /profiles
// 2. Unmarshal the JSON into a simplejson struct
// 3. Ensure the JSON fields are expected
func TestCreateProfile_gateway(t *testing.T) {
	dbTest.Reset(t)
	profile := pb.Profile{
		Name: "work", Notes: "profile for work-related topics",
	}
	resCreate, err := MakeRequestWithDefaults(
		http.MethodPost,
		"http://localhost:8080/v1/profiles",
		profile,
	)
	if err != nil {
		t.Fatalf("unable to create profile: %v", err)
	}
	ValidateResponseCode(t, resCreate, http.StatusOK)
	createJSON, err := simplejson.NewFromReader(resCreate.Body)
	if err != nil {
		t.Fatalf("unable to unmarshal json response: %v", err)
	}
	var tests = []struct {
		name   string
		json   *simplejson.Json
		expect string
	}{
		{
			name:   "profile id",
			json:   createJSON.GetPath("result", "id"),
			expect: `"atlas-contacts-app/profiles/1"`,
		},
		{
			name:   "profile notes",
			json:   createJSON.GetPath("result", "notes"),
			expect: `"profile for work-related topics"`,
		},
		{
			name:   "profile name",
			json:   createJSON.GetPath("result", "name"),
			expect: `"work"`,
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

// TestReadProfile_gateway uses the REST gateway to create a new profile and
// then read that profile from the application by ID
// 1. Create a profile entry with a POST request to /profiles
// 2. Get the profile by sending a GET request to /profiles/{id}
// 2. Unmarshal the JSON into a simplejson struct
// 3. Ensure the JSON fields match what is expected
func TestReadProfile_gateway(t *testing.T) {
	dbTest.Reset(t)
	profile := pb.Profile{
		Name: "personal", Notes: "profile for personal matters",
	}
	resCreate, err := MakeRequestWithDefaults(
		http.MethodPost,
		"http://localhost:8080/v1/profiles",
		profile,
	)
	if err != nil {
		t.Fatalf("unable to create profile: %v", err)
	}
	createJSON, err := simplejson.NewFromReader(resCreate.Body)
	if err != nil {
		t.Fatalf("unable to unmarshal json response: %v", err)
	}
	id, err := createJSON.GetPath("result", "id").String()
	if err != nil {
		t.Fatalf("unable to get profile id from json response: %v", err)
	}
	id = strings.TrimPrefix(id, fmt.Sprintf("%s/%s/", cmd.ApplicationID, "profiles"))
	resRead, err := MakeRequestWithDefaults(
		http.MethodGet, fmt.Sprintf("http://localhost:8080/v1/profiles/%s", id),
		nil,
	)
	if err != nil {
		t.Fatalf("unable to get profile: %v", err)
	}
	ValidateResponseCode(t, resCreate, http.StatusOK)
	readJSON, err := simplejson.NewFromReader(resRead.Body)
	if err != nil {
		t.Fatalf("unable to get profile id from json response: %v", err)
	}
	var tests = []struct {
		name   string
		json   *simplejson.Json
		expect string
	}{
		{
			name:   "profile id",
			json:   readJSON.GetPath("result", "id"),
			expect: `"atlas-contacts-app/profiles/1"`,
		},
		{
			name:   "profile notes",
			json:   readJSON.GetPath("result", "notes"),
			expect: `"profile for personal matters"`,
		},
		{
			name:   "profile name",
			json:   readJSON.GetPath("result", "name"),
			expect: `"personal"`,
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

// TestUpdateProfile_gateway uses the REST gateway to create a new profile and
// then read that profile from the application by ID
// 1. Create a profile with a POST request to /profiles
// 2. Use the corresponding ID to GET the profile at /profiles
// 3. Create a separate profile with updated fields
// 4. Send a PUT request to /profiles/ID with the updated fields
// 5. Ensure the response from PUT has expected fields
func TestUpdateProfile_gateway(t *testing.T) {
	dbTest.Reset(t)
	profile := pb.Profile{
		Name: "photography", Notes: "profile to show my photography portfolio",
	}
	resCreate, err := MakeRequestWithDefaults(
		http.MethodPost,
		"http://localhost:8080/v1/profiles",
		profile,
	)
	if err != nil {
		t.Fatalf("unable to create profile: %v", err)
	}
	createJSON, err := simplejson.NewFromReader(resCreate.Body)
	if err != nil {
		t.Fatalf("unable to unmarshal json response: %v", err)
	}
	id, err := createJSON.GetPath("result", "id").String()
	if err != nil {
		t.Fatalf("unable to get profile id from json response: %v", err)
	}
	id = strings.TrimPrefix(id, fmt.Sprintf("%s/%s/", cmd.ApplicationID, "profiles"))
	updated := pb.Profile{
		Name:  "woodworking",
		Notes: "profile to show my woodworking portfolio",
	}
	resUpdate, err := MakeRequestWithDefaults(
		http.MethodPut, fmt.Sprintf("http://localhost:8080/v1/profiles/%s", id),
		updated,
	)
	if err != nil {
		t.Fatalf("unable to update profile: %v", err)
	}
	updateJSON, err := simplejson.NewFromReader(resUpdate.Body)
	if err != nil {
		t.Fatalf("unable to unmarshal json response: %v", err)
	}
	var tests = []struct {
		name   string
		json   *simplejson.Json
		expect string
	}{
		{
			name:   "profile id",
			json:   updateJSON.GetPath("result", "id"),
			expect: `"atlas-contacts-app/profiles/1"`,
		},
		{
			name:   "profile notes",
			json:   updateJSON.GetPath("result", "notes"),
			expect: `"profile to show my woodworking portfolio"`,
		},
		{
			name:   "profile name",
			json:   updateJSON.GetPath("result", "name"),
			expect: `"woodworking"`,
		},
		{
			name:   "success response",
			json:   updateJSON.GetPath("success"),
			expect: `{"code":"OK","status":200}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ValidateJSONSchema(t, test.json, test.expect)
		})
	}
}

// TestDeleteProfile_gateway uses the REST gateway to create a new profile and
// ensures it can get deleted by sending a DELETE request to /profiles/{id}
// 1. Create a profile entry with a POST request
// 2. Delete the profile by sending a DELETE request to /profiles/{id}
// 3. Ensure the DELETE response JSON matches the expected schema
func TestDeleteProfile_gateway(t *testing.T) {
	dbTest.Reset(t)
	profile := pb.Profile{
		Name: "school", Notes: "profile for academic-related content",
	}
	resCreate, err := MakeRequestWithDefaults(
		http.MethodPost,
		"http://localhost:8080/v1/profiles",
		profile,
	)
	if err != nil {
		t.Fatalf("unable to create profile: %v", err)
	}
	createJSON, err := simplejson.NewFromReader(resCreate.Body)
	if err != nil {
		t.Fatalf("unable to unmarshal json response: %v", err)
	}
	id, err := createJSON.GetPath("result", "id").String()
	if err != nil {
		t.Fatalf("unable to get profile id from json response: %v", err)
	}
	id = strings.TrimPrefix(id, fmt.Sprintf("%s/%s/", cmd.ApplicationID, "profiles"))
	resDelete, err := MakeRequestWithDefaults(
		http.MethodDelete,
		fmt.Sprintf("http://localhost:8080/v1/profiles/%s", id),
		nil,
	)
	if err != nil {
		t.Fatalf("unable to delete profile: %v", err)
	}
	ValidateResponseCode(t, resDelete, http.StatusOK)
	deleteJSON, err := simplejson.NewFromReader(resDelete.Body)
	if err != nil {
		t.Fatalf("unable to unmarshal json response: %v", err)
	}
	t.Run("success response", func(t *testing.T) {
		ValidateJSONSchema(t, deleteJSON, `{"success":{"code":"OK","status":200}}`)
	})
}

// TestListProfiles_gateway uses the REST gateway to create two profiles and
// both profiles are included in a GET request to /profiles
// 1. Create two profiles with POST requests to /profiles
// 2. Send a GET request to /profiles to get both profiles
// 3. Validate the response JSON to ensure it includes two profiles
func TestListProfiles_gateway(t *testing.T) {
	dbTest.Reset(t)
	first := pb.Profile{
		Name: "cooking", Notes: "profile for cooking projects",
	}
	second := pb.Profile{
		Name: "family", Notes: "profile for family information",
	}
	if _, err := MakeRequestWithDefaults(
		http.MethodPost,
		"http://localhost:8080/v1/profiles",
		first,
	); err != nil {
		t.Fatalf("unable to create first profile %v", err)
	}
	if _, err := MakeRequestWithDefaults(
		http.MethodPost,
		"http://localhost:8080/v1/profiles",
		second,
	); err != nil {
		t.Fatalf("unable to create second profile %v", err)
	}
	resList, err := MakeRequestWithDefaults(
		http.MethodGet,
		"http://localhost:8080/v1/profiles",
		nil,
	)
	if err != nil {
		t.Fatalf("unable to list profiles %v", err)
	}
	ValidateResponseCode(t, resList, http.StatusOK)
	listJSON, err := simplejson.NewFromReader(resList.Body)
	if err != nil {
		t.Fatalf("unable to unmarshal json response: %v", err)
	}
	var tests = []struct {
		name   string
		json   *simplejson.Json
		expect string
	}{
		{
			name:   "first profile",
			json:   listJSON.Get("results").GetIndex(0),
			expect: `{"id":"atlas-contacts-app/profiles/1","name":"cooking","notes":"profile for cooking projects"}`,
		},
		{
			name:   "second profile",
			json:   listJSON.Get("results").GetIndex(1),
			expect: `{"id":"atlas-contacts-app/profiles/2","name":"family","notes":"profile for family information"}`,
		},
		{
			name:   "success response",
			json:   listJSON.GetPath("success"),
			expect: `{"code":"OK","status":200}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ValidateJSONSchema(t, test.json, test.expect)
		})
	}
}
