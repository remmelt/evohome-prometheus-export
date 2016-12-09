package main

import (
	"fmt"
	"github.com/jcmturner/evohome-prometheus-export/authenticate"
	"github.com/jcmturner/evohome-prometheus-export/handlers"
	"github.com/jcmturner/evohome-prometheus-export/installation"
	"github.com/jcmturner/evohome-prometheus-export/location"
	"github.com/jcmturner/evohome-prometheus-export/logging"
	"github.com/jcmturner/evohome-prometheus-export/userAccount"
	"github.com/jcmturner/restclient"
	"net/http"
	"os"
)

var githash = "No version available"
var buildstamp = "Not set"

const (
	HTTPPort        = 8080
	ServiceEndPoint = "https://tccna.honeywell.com"
)

func main() {
	c := restclient.NewConfig()
	c.WithEndPoint(ServiceEndPoint)
	c.WithCAFilePath(os.Getenv("TRUST_CERT"))
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

	logs.Info.Printf(`EvoHome to Prometheus - Configuration Complete:
	Build hash: %s
	Build timestap: %s
	Listenning Port: %v
	Service URL: %s
	CA Trust Path: %s`, githash, buildstamp, HTTPPort, ServiceEndPoint, os.Getenv("TRUST_CERT"))

	err = http.ListenAndServe(fmt.Sprintf(":%v", HTTPPort), mux)
	logs.Error.Fatalf("HTTP Server Exit: %v\n", err)
}
