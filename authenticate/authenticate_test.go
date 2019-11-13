package authenticate

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/jcmturner/restclient"
	"github.com/remmelt/evohome-prometheus-export/logging"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	userId          = "b013aa26-9724-4dbd-8897-048b9aada249"
	password        = "test"
	evohomeUid      = "username@example.com"
	evohomePassword = "somepassword"
	responseData    = `{
  "access_token": "test-access-token",
  "token_type": "bearer",
  "expires_in": 3599,
  "refresh_token": "test-refresh-token",
  "scope": "EMEA-V1-Anonymous"
}`
)

func checkAuth(r *http.Request) bool {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return false
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}
	return pair[0] == userId && pair[1] == password
}

func checkAuthPost(r *http.Request) bool {
	if r.Method != "POST" {
		return false
	}
	r.ParseForm()
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	v, _ := url.ParseQuery(string(body))
	if v.Get("Content-Type") == "application/x-www-form-urlencoded; charset=utf-8" &&
		v.Get("Cache-Control") == "no-store no-cache" &&
		v.Get("Pragma") == "no-cache" &&
		v.Get("grant_type") == "password" &&
		v.Get("scope") == "EMEA-V1-Basic EMEA-V1-Anonymous EMEA-V1-Get-Current-User-Account" &&
		v.Get("Username") == evohomeUid &&
		v.Get("Password") == evohomePassword {
		return true
	} else {
		return false
	}
}

func testServer() *httptest.Server {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) {
			if checkAuthPost(r) {
				w.Header().Set("Content-Type", "application/json;charset=UTF-8")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Expires", "-1")
				w.Header().Set("Pragma", "no-cache")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, responseData)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	return s
}

func TestAuthenticate(t *testing.T) {
	os.Setenv("EVOHOME_USERNAME", evohomeUid)
	os.Setenv("EVOHOME_PASSWORD", evohomePassword)
	s := testServer()
	//Get certifcate from test TLS server, output in PEM format to file
	certOut, _ := ioutil.TempFile(os.TempDir(), "testCert")
	defer os.Remove(certOut.Name())
	certBytes := s.TLS.Certificates[0].Certificate[0]
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

	c := restclient.NewConfig()
	c.WithEndPoint(s.URL)
	c.WithCAFilePath(certOut.Name())
	logs, _ := logging.LoggerSetUp()

	var a Authenticate
	err := a.NewRequest(c, logs)
	if err != nil {
		t.Errorf("Could not prepare authentication request: %v\n", err)
	}
	err = a.Process()
	if err != nil {
		t.Fatalf("Error processing request: %s", err)
	}
	assert.Equal(t, "test-access-token", a.AccessToken, "Access token not set as expected")
	assert.Equal(t, "bearer test-access-token", a.IdentityHeaders.Authorization, "Authorization details not set as expected")
	assert.Equal(t, userId, a.IdentityHeaders.ApplicationID, "ApplicationID not set as expected")

	//Test usng a cached token. Manually change the token value and check it is not updated
	a.AccessToken = "cached_token"
	err = a.Process()
	if err != nil {
		t.Errorf("Error processing request: %s", err)
	}
	assert.Equal(t, "cached_token", a.AccessToken, "The cached token was not used")

	//Test renewal. Manually update the validUntil value and check the token is not updated
	a.AccessToken = "cached_token"
	a.validUntil = time.Now().Add(time.Duration(-10) * time.Second)
	err = a.Process()
	if err != nil {
		t.Errorf("Error processing request: %s", err)
	}
	assert.Equal(t, "test-access-token", a.AccessToken, "Access token has not been renewed")
}
