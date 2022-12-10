package shortdescription_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"golang.org/x/sync/errgroup"

	shortdescription "github.com/Inuart/wikimedia-exercise"
)

func TestDescriptorHandler(t *testing.T) {
	var mockClient mockHttpClient

	descriptor, err := shortdescription.New(shortdescription.Config{
		ContactInfo: "test case",
		HttpClient:  &mockClient,
		CachedTTL:   -1, // remove caching
	})
	if err != nil {
		t.Fatal(err)
	}

	client := startTestServer(t, descriptor)

	testCases := []struct {
		name             string
		person           string
		upstreamCode     int
		upstreamResponse wikiJSON
		expectedCode     int
		expectedResult   string
	}{
		{
			name:         "person param missing",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "upstream error",
			person:       testPerson,
			upstreamCode: http.StatusInternalServerError,
			expectedCode: http.StatusBadGateway,
		},
		{
			name:             "check result",
			person:           testPerson,
			upstreamResponse: testContent,
			expectedResult:   testDescription,
		},
		{
			name:             "check url-encoded input",
			person:           testUrlEncodedPerson,
			upstreamResponse: testContent,
			expectedResult:   testDescription,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient.code = tc.upstreamCode
			mockClient.body = tc.upstreamResponse

			res := client.get(t, tc.person)
			defer res.Body.Close()

			if tc.expectedCode == 0 {
				tc.expectedCode = http.StatusOK
			}

			if res.StatusCode != tc.expectedCode {
				t.Fatalf("wanted %v, got %v: %v", tc.expectedCode, res.StatusCode, responseError(res))
			}

			if res.StatusCode != http.StatusOK {
				return
			}

			var result shortdescription.ShortDescription

			if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
				t.Fatal("json decoding failed", err)
			}

			person, err := url.QueryUnescape(tc.person)
			if err != nil {
				t.Fatal("test case person is wrongly encoded", err)
			}

			if result.Person != person {
				t.Errorf("wanted %v, got %v", person, result.Person)
			}

			if result.Description != tc.expectedResult {
				t.Errorf("wanted %v, got %v", tc.expectedResult, result.Description)
			}
		})
	}
}

// Keeping integration test cases at a minimum to not overload the real service.
func TestDescriptorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests")
	}

	descriptor, err := shortdescription.New(shortdescription.Config{
		ContactInfo: "test case",
	})
	if err != nil {
		t.Fatal(err)
	}

	client := startTestServer(t, descriptor)

	testCases := []struct {
		name           string
		person         string
		expectedCode   int
		expectedResult string
	}{
		{
			name:         "person does not exist",
			person:       "unknown person",
			expectedCode: http.StatusNotFound,
		},
		{
			name:           "check known result",
			person:         testPerson,
			expectedResult: testDescription,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := client.get(t, tc.person)
			defer res.Body.Close()

			if tc.expectedCode == 0 {
				tc.expectedCode = http.StatusOK
			}

			if res.StatusCode != tc.expectedCode {
				t.Fatalf("wanted %v, got %v: %v", tc.expectedCode, res.StatusCode, responseError(res))
			}

			if res.StatusCode != http.StatusOK {
				return
			}

			var result shortdescription.ShortDescription

			if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
				t.Fatal("json decoding failed", err)
			}

			if result.Person != tc.person {
				t.Errorf("wanted %v, got %v", tc.person, result.Person)
			}

			if result.Description != tc.expectedResult {
				t.Errorf("wanted %v, got %v", tc.expectedResult, result.Description)
			}
		})
	}
}

func TestDescriptorHandlerConcurrently(t *testing.T) {
	descriptor, err := shortdescription.New(shortdescription.Config{
		ContactInfo: testContactInfo,
		CachedTTL:   -1,
	})
	if err != nil {
		t.Fatal(err)
	}

	client := startTestServer(t, descriptor)

	const concurrentRequests = 99

	var eg errgroup.Group

	for i := 0; i < concurrentRequests; i++ {
		eg.Go(func() error {
			res := client.get(t, testPerson)
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				return fmt.Errorf("wanted %v, got %v: %v", http.StatusOK, res.StatusCode, responseError(res))
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
}
