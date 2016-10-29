package installation

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"io/ioutil"
	"os"
	"encoding/pem"
	"github.com/jcmturner/restclient"
	"github.com/jcmturner/evohome-prometheus-export/logging"
	"github.com/jcmturner/evohome-prometheus-export/authenticate"
	"github.com/stretchr/testify/assert"
)

const (
	accessToken = "bearer test-access-token"
	userId = "1234567"
	evohomeUid      = "username@example.com"
	evohomePassword = "somepassword"
	authResponseData = `{
  "access_token": "test-access-token",
  "token_type": "bearer",
  "expires_in": 3599,
  "refresh_token": "test-refresh-token",
  "scope": "EMEA-V1-Anonymous"
}`
	responseData = `[
  {
    "locationInfo": {
      "locationId": "1234567",
      "name": "Home",
      "streetAddress": "10 Downing St",
      "city": "LONDON",
      "country": "UnitedKingdom",
      "postcode": "SW1A 2AA",
      "locationType": "Residential",
      "useDaylightSaveSwitching": true,
      "timeZone": {
        "timeZoneId": "GMTStandardTime",
        "displayName": "(UTC) Dublin, Edinburgh, Lisbon, London",
        "offsetMinutes": 0,
        "currentOffsetMinutes": 60,
        "supportsDaylightSaving": true
      },
      "locationOwner": {
        "userId": "1234567",
        "username": "username@example.com",
        "firstname": "fname",
        "lastname": "lname"
      }
    },
    "gateways": [
      {
        "gatewayInfo": {
          "gatewayId": "1234567",
          "mac": "00D012AD23DF",
          "crc": "2ED4",
          "isWiFi": false
        },
        "temperatureControlSystems": [
          {
            "systemId": "2345678",
            "modelType": "EvoTouch",
            "zones": [
              {
                "zoneId": "1234567",
                "modelType": "HeatingZone",
                "heatSetpointCapabilities": {
                  "maxHeatSetpoint": 35,
                  "minHeatSetpoint": 5,
                  "valueResolution": 0.5,
                  "allowedSetpointModes": [
                    "PermanentOverride",
                    "FollowSchedule",
                    "TemporaryOverride"
                  ],
                  "maxDuration": "1.00:00:00",
                  "timingResolution": "00:10:00"
                },
                "scheduleCapabilities": {
                  "maxSwitchpointsPerDay": 6,
                  "minSwitchpointsPerDay": 1,
                  "timingResolution": "00:10:00",
                  "setpointValueResolution": 0.5
                },
                "name": "Radiators",
                "zoneType": "ZoneTemperatureControl"
              },
              {
                "zoneId": "1234568",
                "modelType": "HeatingZone",
                "heatSetpointCapabilities": {
                  "maxHeatSetpoint": 35,
                  "minHeatSetpoint": 5,
                  "valueResolution": 0.5,
                  "allowedSetpointModes": [
                    "PermanentOverride",
                    "FollowSchedule",
                    "TemporaryOverride"
                  ],
                  "maxDuration": "1.00:00:00",
                  "timingResolution": "00:10:00"
                },
                "scheduleCapabilities": {
                  "maxSwitchpointsPerDay": 6,
                  "minSwitchpointsPerDay": 1,
                  "timingResolution": "00:10:00",
                  "setpointValueResolution": 0.5
                },
                "name": "Kitchen",
                "zoneType": "ZoneTemperatureControl"
              }
            ],
            "allowedSystemModes": [
              {
                "systemMode": "Auto",
                "canBePermanent": true,
                "canBeTemporary": false
              },
              {
                "systemMode": "AutoWithEco",
                "canBePermanent": true,
                "canBeTemporary": true,
                "maxDuration": "1.00:00:00",
                "timingResolution": "01:00:00",
                "timingMode": "Duration"
              },
              {
                "systemMode": "AutoWithReset",
                "canBePermanent": true,
                "canBeTemporary": false
              },
              {
                "systemMode": "Away",
                "canBePermanent": true,
                "canBeTemporary": true,
                "maxDuration": "99.00:00:00",
                "timingResolution": "1.00:00:00",
                "timingMode": "Period"
              },
              {
                "systemMode": "DayOff",
                "canBePermanent": true,
                "canBeTemporary": true,
                "maxDuration": "99.00:00:00",
                "timingResolution": "1.00:00:00",
                "timingMode": "Period"
              },
              {
                "systemMode": "HeatingOff",
                "canBePermanent": true,
                "canBeTemporary": false
              },
              {
                "systemMode": "Custom",
                "canBePermanent": true,
                "canBeTemporary": true,
                "maxDuration": "99.00:00:00",
                "timingResolution": "1.00:00:00",
                "timingMode": "Period"
              }
            ]
          }
        ]
      }
    ]
  }
]`
)

func checkAuth(r *http.Request) bool {
	if r.Header.Get("Authorization") == accessToken {
		return true
	}
	return false
}

func checkQueryData(r *http.Request) bool {
	v := r.URL.Query()
	if v.Get("includeTemperatureControlSystems") == "True" && v.Get("userId") != "" {
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

func TestInstallation(t *testing.T) {
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

	var i Installation
	err = i.NewRequest(userId, c, logs)
	if err != nil {
		t.Fatalf("Could not prepare Installation request: %v\n", err)
	}
	locationID, err := i.GetLocationID(&a)
	if err != nil {
		t.Errorf("Failed to get location ID: %v\n", err)
	}
	assert.Equal(t, "1234567", locationID, "Location ID not as expected")
	systemID, err := i.GetSystemID(&a)
	if err != nil {
		t.Errorf("Failed to get system ID: %v\n", err)
	}
	assert.Equal(t, "2345678", systemID, "System ID not as expected")
	zones, err := i.GetTemperatureControlSystemZones(&a)
	if err != nil {
		t.Errorf("Failed to get temperature control zones: %v\n", err)
	}
	assert.True(t, len(zones) > 0, "Did not get any zones")
}
