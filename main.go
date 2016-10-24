package main

import (
	"github.com/jcmturner/restclient"
	"os"
	"github.com/jcmturner/evohome-prometheus-export/authenticate"
	"github.com/jcmturner/evohome-prometheus-export/userAccount"
	"github.com/jcmturner/evohome-prometheus-export/installation"
	"github.com/jcmturner/evohome-prometheus-export/location"
	"fmt"
)

func main() {
	c := restclient.NewConfig()
	c.WithEndPoint("https://tccna.honeywell.com")
	c.WithCAFilePath(os.Getenv("TRUST_CERT"))

	var a authenticate.Authenticate
	err := a.NewRequest(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Could not prepare authentication request: %v\n", err)
		os.Exit(1)
	}

	var u userAccount.UserAccount
	err = u.NewRequest(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Could not prepare userAccount request: %v\n", err)
		os.Exit(1)
	}
	uid, err := u.GetUserID(&a)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Could not get UserID: %v\n", err)
		os.Exit(1)
	}

	var i installation.Installation
	err = i.NewRequest(uid, c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Could not prepare installation request: %v\n", err)
		os.Exit(1)
	}
	lid, err := i.GetLocationID(&a)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Could not get LocationID: %v\n", err)
		os.Exit(1)
	}

	var l location.Location
	err = l.NewRequest(lid, c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Could not prepare location request: %v\n", err)
		os.Exit(1)
	}
	zones, err := l.GetTemperatureControlSystemZonesStatus(&a)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Could not get zone information: %v\n", err)
		os.Exit(1)
	}

	for _, z := range zones {
		fmt.Printf("Zone: %v\n", z.Name)
		fmt.Printf("Target: %v\n", z.TargetTemperature)
		fmt.Printf("Current: %v\n\n", z.CurrentTemperature)
	}

}