package shortdescription_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	shortdescription "github.com/Inuart/wikimedia-exercise"
)

type testClient struct {
	*http.Client
	url string
}

func startTestServer(t *testing.T, d shortdescription.Describer) testClient {
	server := httptest.NewServer(d)

	t.Cleanup(func() {
		server.Close()
	})

	return testClient{Client: server.Client(), url: server.URL}
}

func (client testClient) get(t *testing.T, person string) *http.Response {
	t.Helper()

	res, err := client.Get(client.url + "?person=" + url.QueryEscape(person))
	if err != nil {
		t.Fatal(err)
	}

	return res
}

func responseError(r *http.Response) error {
	if r.StatusCode < http.StatusBadRequest { // a status code >= 400 is an error
		return nil
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf(
			"request failed with a %q (also, body could not be read: %v)",
			http.StatusText(r.StatusCode), err,
		)
	}

	return fmt.Errorf(
		"request failed with error %q (%s)",
		body, http.StatusText(r.StatusCode),
	)
}

type mockHttpClient struct {
	body wikiJSON
	code int
}

const (
	testPerson             = "Yoshua Bengio"
	testNonCanonicalPerson = "yoshua_Bengio"
	testUrlEncodedPerson   = "Yoshua%20Bengio"
	testBadUrlEncoding     = "Yoshua% Bengio"
	testUserAgent          = "test user agent"
	testContactInfo        = "test contact info"
	testDescription        = "Canadian computer scientist"
	testContent            = "...{{Short description|" + testDescription + "}}..."
)

func (m mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()

	if m.code > 0 {
		w.WriteHeader(m.code)
		return w.Result(), nil
	}

	if m.body != "" {
		_, err := w.WriteString(string(m.body))
		return w.Result(), err
	}

	titles := req.URL.Query().Get("titles")

	person, err := url.QueryUnescape(titles)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.WriteString(person + " has bad encoding")
		return w.Result(), err
	}

	if person == testPerson {
		_, err := w.WriteString(string(responseSample))
		return w.Result(), err
	}

	w.WriteHeader(http.StatusNotFound)
	return w.Result(), nil
}
