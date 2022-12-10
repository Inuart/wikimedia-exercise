package shortdescription_test

import (
	"context"
	"errors"
	"testing"

	shortdescription "github.com/Inuart/wikimedia-exercise"
)

func TestNewDescriptor(t *testing.T) {
	testCases := []struct {
		name      string
		cfg       shortdescription.Config
		expectErr bool
	}{
		{
			name: "happy path at creating a new descriptor",
			cfg: shortdescription.Config{
				ContactInfo: "testKey",
			},
		},
		{
			name:      "fail at creating a new descriptor without an API key",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := shortdescription.New(tc.cfg)
			if tc.expectErr {
				if err == nil {
					t.Error("an error was expected")
				}
			} else if err != nil {
				t.Error("unexpected error", err)
			}
		})
	}
}

func TestDescriptorClient(t *testing.T) {
	ctx := context.Background()
	var mockClient mockHttpClient

	descriptor, err := shortdescription.New(shortdescription.Config{
		ContactInfo: testContactInfo,
		HttpClient:  &mockClient,
	})
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name        string
		person      string
		userAgent   string
		response    wikiJSON
		expected    string
		expectedErr error
	}{
		{
			name:        "fails if there's no person argument",
			userAgent:   testUserAgent,
			expectedErr: shortdescription.ErrInvalidArgument,
		},
		{
			name:        "fails if there's no userAgent argument",
			person:      testPerson,
			expectedErr: shortdescription.ErrInvalidArgument,
		},
		{
			name:        "fails if there's no short description",
			person:      testPerson,
			userAgent:   testUserAgent,
			response:    "{{Long description|Canadian computer scientist}}",
			expectedErr: shortdescription.ErrNotFound,
		},
		{
			name:      "happy path returns a short description",
			person:    testPerson,
			userAgent: testUserAgent,
			response:  responseSample,
			expected:  "Canadian computer scientist",
		},
		{
			name:        "fails if the person's description is not found",
			person:      "unknown person",
			userAgent:   testUserAgent,
			expectedErr: shortdescription.ErrNotFound,
		},
		{
			name:      "returns a short description from the cache",
			person:    testPerson,
			userAgent: testUserAgent,
			expected:  "Canadian computer scientist",
		},
		{
			name:      "normalizes the input and returns a short description from the cache",
			person:    testNonCanonicalPerson,
			userAgent: testUserAgent,
			expected:  "Canadian computer scientist",
		},
		{
			name:      "url-encoded input returns a short description from the cache",
			person:    testUrlEncodedPerson,
			userAgent: testUserAgent,
			expected:  "Canadian computer scientist",
		},
		{
			name:        "wrongly url-encoded input returns an error",
			person:      testBadUrlEncoding,
			userAgent:   testUserAgent,
			expectedErr: shortdescription.ErrInvalidArgument,
		},
		{
			name:      "avoids multiple persons in the same query",
			person:    testPerson + "|France",
			userAgent: testUserAgent,
			expected:  "Canadian computer scientist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient.body = tc.response

			descr, err := descriptor.ShortDescription(ctx, tc.person, tc.userAgent)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("wanted %v, got %v", tc.expectedErr, err)
			}

			if descr.Description != tc.expected {
				t.Errorf("wanted %s, got %s", tc.expected, descr.Description)
			}
		})
	}
}
