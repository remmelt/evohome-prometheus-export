package userAccount

import (
	"io/ioutil"
	"os"
	"encoding/pem"
	"github.com/jcmturner/restclient"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"fmt"
	"github.com/jcmturner/evohome-prometheus-export/logging"
	"net/http"
	"github.com/jcmturner/evohome-prometheus-export/authenticate"
	"testing"
)

const (
	accessToken = "bearer test-access-token"
	evohomeUid      = "username@example.com"
	evohomePassword = "somepassword"
	authResponseData = `{
  "access_token": "test-access-token",
  "token_type": "bearer",
  "expires_in": 3599,
  "refresh_token": "test-refresh-token",
  "scope": "EMEA-V1-Anonymous"
}`
	responseData = `{
  "userId": "1234567",
  "username": "username@example.com",
  "firstname": "fname",
  "lastname": "lname",
  "streetAddress": "10 Downing St",
  "city": "LONDON",
  "postcode": "SW1A 2AA",
  "country": "UnitedKingdom",
  "language": "enGB"
}`
)

func checkAuth(r *http.Request) bool {
	if r.Header.Get("Authorization") == accessToken {
		return true
	}
	return false
}

func testAuthServer() *httptest.Server {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Expires", "-1")
		w.Header().Set("Pragma", "no-cache")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, authResponseData)
	}))
	return s
}

func testServer() *httptest.Server {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) {
				w.Header().Set("Content-Type", "application/json;charset=UTF-8")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Expires", "-1")
				w.Header().Set("Pragma", "no-cache")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, responseData)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	return s
}

func TestUserAccount(t *testing.T) {
	os.Setenv("EVOHOME_USERNAME", evohomeUid)
	os.Setenv("EVOHOME_PASSWORD", evohomePassword)
	os.Setenv("LOG_LEVEL", "DEBUG")
	as := testAuthServer()
	//Get certifcate from test TLS server, output in PEM format to file
	certOut, _ := ioutil.TempFile(os.TempDir(), "testCert")
	defer os.Remove(certOut.Name())
	certBytes := as.TLS.Certificates[0].Certificate[0]
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

	c := restclient.NewConfig()
	c.WithEndPoint(as.URL)
	c.WithCAFilePath(certOut.Name())
	logs, _ := logging.LoggerSetUp()

	var a authenticate.Authenticate
	err := a.NewRequest(c, logs)
	if err != nil {
		t.Fatalf("Could not prepare authentication request: %v\n", err)
	}
	err = a.Process()
	if err != nil {
		t.Fatalf("Error processing request: %s", err)
	}

	s := testServer()
	//Get certifcate from test TLS server, output in PEM format to file
	certOut, _ = ioutil.TempFile(os.TempDir(), "testCert")
	defer os.Remove(certOut.Name())
	certBytes = s.TLS.Certificates[0].Certificate[0]
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

	c = restclient.NewConfig()
	c.WithEndPoint(s.URL)
	c.WithCAFilePath(certOut.Name())

	var u UserAccount
	err = u.NewRequest(c, logs)
	if err != nil {
		t.Fatalf("Could not prepare UserAccount request: %v\n", err)
	}

	uid, err := u.GetUserID(&a)
	if err != nil {
		t.Fatalf("Could not get userID: %v\n", err)
	}
	assert.Equal(t, "1234567", uid, "UserID not as expected")
	city, err := u.GetCity(&a)
	if err != nil {
		t.Fatalf("Could not get City: %v\n", err)
	}
	assert.Equal(t, "LONDON", city, "City not as expected")
	country, err := u.GetCountry(&a)
	if err != nil {
		t.Fatalf("Could not get country: %v\n", err)
	}
	assert.Equal(t, "UnitedKingdom", country, "Country not as expected")
	fn, err := u.GetFirstname(&a)
	if err != nil {
		t.Fatalf("Could not get first name: %v\n", err)
	}
	assert.Equal(t, "fname", fn, "First name not as expected")
	ln, err := u.GetLastname(&a)
	if err != nil {
		t.Fatalf("Could not get last name: %v\n", err)
	}
	assert.Equal(t, "lname", ln, "Last name not as expected")
	lang, err := u.GetLanguage(&a)
	if err != nil {
		t.Fatalf("Could not get language: %v\n", err)
	}
	assert.Equal(t, "enGB", lang, "Language not as expected")
	pc, err := u.GetPostcode(&a)
	if err != nil {
		t.Fatalf("Could not get postcode: %v\n", err)
	}
	assert.Equal(t, "SW1A 2AA", pc, "Postcode not as expected")
	sa, err := u.GetStreetAddress(&a)
	if err != nil {
		t.Fatalf("Could not get street address: %v\n", err)
	}
	assert.Equal(t, "10 Downing St", sa, "Street address not as expected")
	uname, err := u.GetUsername(&a)
	if err != nil {
		t.Fatalf("Could not get username: %v\n", err)
	}
	assert.Equal(t, "username@example.com", uname, "Username not as expected")
}
