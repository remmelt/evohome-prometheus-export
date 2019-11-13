package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/jcmturner/restclient"
	"github.com/remmelt/evohome-prometheus-export/authenticate"
	"github.com/remmelt/evohome-prometheus-export/handlers"
	"github.com/remmelt/evohome-prometheus-export/installation"
	"github.com/remmelt/evohome-prometheus-export/location"
	"github.com/remmelt/evohome-prometheus-export/logging"
	"github.com/remmelt/evohome-prometheus-export/userAccount"
)

var githash = "No version available"
var buildstamp = "Not set"

const (
	serviceEndPoint = "https://tccna.honeywell.com"
)

func main() {
	c := restclient.NewConfig()
	c.WithEndPoint(serviceEndPoint)
	certPath := os.Getenv("TRUST_CERT")
	c.TrustCACert = &certPath
	c.WithCAFilePath(certPath)

	if err := c.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Configuration of web service not valid: %v", err)
		os.Exit(1)
	}

	logs, err := logging.LoggerSetUp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Could not set up logging: %v", err)
	}

	var a authenticate.Authenticate
	err = a.NewRequest(c, logs)
	if err != nil {
		logs.Error.Fatalf("Could not prepare authentication request: %v\n", err)
	}

	var u userAccount.UserAccount
	err = u.NewRequest(c, logs)
	if err != nil {
		logs.Error.Fatalf("Could not prepare userAccount request: %v\n", err)
	}
	uid, err := u.GetUserID(&a)
	if err != nil {
		logs.Error.Fatalf("Could not get UserID: %v\n", err)
	}

	var i installation.Installation
	err = i.NewRequest(uid, c, logs)
	if err != nil {
		logs.Error.Fatalf("Could not prepare installation request: %v\n", err)
	}
	lid, err := i.GetLocationID(&a)
	if err != nil {
		logs.Error.Fatalf("Could not get LocationID: %v\n", err)
	}

	var l location.Location
	err = l.NewRequest(lid, c, logs)
	if err != nil {
		logs.Error.Fatalf("Could not prepare location request: %v\n", err)
	}

	//Set up handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/zoneTemperatures", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetZoneTemperatures(w, &a, &l, logs)
	})

	httpPort := getEnv("SERVER_PORT", "8080")
	logs.Info.Printf(`EvoHome to Prometheus - Configuration Complete:
	Build hash: %s
	Build timestap: %s
	Listening Port: %s
	Service URL: %s
	CA Trust Path: %s`, githash, buildstamp, httpPort, serviceEndPoint, certPath)

	err = http.ListenAndServe(fmt.Sprintf(":%v", httpPort), mux)
	logs.Error.Fatalf("HTTP Server Exit: %v\n", err)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
