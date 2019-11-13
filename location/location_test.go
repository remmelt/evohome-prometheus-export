package location

import (
	"encoding/pem"
	"fmt"
	"github.com/jcmturner/restclient"
	"github.com/remmelt/evohome-prometheus-export/authenticate"
	"github.com/remmelt/evohome-prometheus-export/logging"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const (
	accessToken      = "bearer test-access-token"
	locationId       = "1234567"
	evohomeUid       = "username@example.com"
	evohomePassword  = "somepassword"
	authResponseData = `{
  "access_token": "test-access-token",
  "token_type": "bearer",
  "expires_in": 3599,
  "refresh_token": "test-refresh-token",
  "scope": "EMEA-V1-Anonymous"
}`
	responseData = `{
  "locationId": "1234567",
  "gateways": [
    {
      "gatewayId": "1234567",
      "temperatureControlSystems": [
        {
          "systemId": "1234567",
          "zones": [
            {
              "zoneId": "1234567",
              "temperatureStatus": {
                "temperature": 22.5,
                "isAvailable": true
              },
              "activeFaults": [],
              "heatSetpointStatus": {
                "targetTemperature": 22,
                "setpointMode": "FollowSchedule"
              },
              "name": "Radiators"
            },
            {
              "zoneId": "2345678",
              "temperatureStatus": {
                "temperature": 22.5,
                "isAvailable": true
              },
              "activeFaults": [],
              "heatSetpointStatus": {
                "targetTemperature": 22,
                "setpointMode": "FollowSchedule"
              },
              "name": "Kitchen"
            }
          ],
          "activeFaults": [],
          "systemModeStatus": {
            "mode": "Auto",
            "isPermanent": true
          }
        }
      ],
      "activeFaults": []
    }
  ]
}`
)

func checkAuth(r *http.Request) bool {
	if r.Header.Get("Authorization") == accessToken {
		return true
	}
	return false
}

func checkQueryData(r *http.Request) bool {
	v := r.URL.Query()
	if v.Get("includeTemperatureControlSystems") == "True" {
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
			if checkQueryData(r) {
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

func TestLocation(t *testing.T) {
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

	var l Location
	err = l.NewRequest(locationId, c, logs)
	if err != nil {
		t.Fatalf("Could not prepare Location request: %v\n", err)
	}
	zones, err := l.GetTemperatureControlSystemZonesStatus(&a)
	if err != nil {
		t.Fatalf("Could not get temperature control system zones status: %v\n", err)
	}
	assert.True(t, len(zones) > 0, "Did not get any zones")
	assert.Equal(t, float32(22.5), zones[0].CurrentTemperature, "Current zone temperature not as expected")
}
