// +build integration

package integration

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	simplejson "github.com/bitly/go-simplejson"
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
	if e := http.StatusOK; resCreate.StatusCode != e {
		t.Errorf("got: %d wanted: %d", resCreate.StatusCode, http.StatusOK)
	}

	bs, err := ioutil.ReadAll(resCreate.Body)
	if err != nil {
		t.Fatal(err)
	}

	e := `{"result":{"id":"atlas-contacts-app/profile/1","name":"work","notes":"profile for work-related topics"}}`

	if body := string(bs); e != body {
		t.Errorf("got: %s wanted: %s", body, e)
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
	parts := strings.Split(id, "/")
	id = parts[len(parts)-1]
	resRead, err := MakeRequestWithDefaults(
		http.MethodGet, fmt.Sprintf("http://localhost:8080/v1/profiles/%s", id),
		nil,
	)
	if err != nil {
		t.Fatalf("unable to get profile: %v", err)
	}

	ValidateResponseCode(t, resCreate, http.StatusOK)

	bs, err := ioutil.ReadAll(resRead.Body)
	if err != nil {
		t.Fatal(err)
	}

	e := `{"result":{"id":"atlas-contacts-app/profile/1","name":"personal","notes":"profile for personal matters"}}`

	if body := string(bs); e != body {
		t.Errorf("got: %s wanted: %s", body, e)
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
	parts := strings.Split(id, "/")
	id = parts[len(parts)-1]
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
			expect: `"atlas-contacts-app/profile/1"`,
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
	parts := strings.Split(id, "/")
	id = parts[len(parts)-1]
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

	ValidateJSONSchema(t, deleteJSON, `{}`)
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

	if e := http.StatusOK; resList.StatusCode != e {
		t.Errorf("got: %d wanted: %d", resList.StatusCode, http.StatusOK)
	}

	bs, err := ioutil.ReadAll(resList.Body)
	if err != nil {
		t.Fatal(err)
	}

	e := `{"results":[{"id":"atlas-contacts-app/profile/1","name":"cooking","notes":"profile for cooking projects"},{"id":"atlas-contacts-app/profile/2","name":"family","notes":"profile for family information"}]}`

	if body := string(bs); e != body {
		t.Errorf("got: %s wanted: %s", body, e)
	}

}
