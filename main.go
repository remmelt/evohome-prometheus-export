package main

import (
	"github.com/jcmturner/restclient"
	"github.com/jcmturner/evohome-prometheus-export/authenticate"
	"github.com/jcmturner/evohome-prometheus-export/userAccount"
	"github.com/jcmturner/evohome-prometheus-export/installation"
	"github.com/jcmturner/evohome-prometheus-export/location"
	"fmt"
	"os"
	"github.com/jcmturner/evohome-prometheus-export/logging"
)

func main() {
	c := restclient.NewConfig()
	c.WithEndPoint("https://tccna.honeywell.com")
	c.WithCAFilePath(os.Getenv("TRUST_CERT"))

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
	zones, err := l.GetTemperatureControlSystemZonesStatus(&a)
	if err != nil {
		logs.Error.Fatalf("Could not get zone information: %v\n", err)
	}

	for _, z := range zones {
		logs.Info.Printf("Zone: %v\n", z.Name)
		logs.Info.Printf("Target: %v\n", z.TargetTemperature)
		logs.Info.Printf("Current: %v\n\n", z.CurrentTemperature)
	}

}

