package authenticate

import (
	"errors"
	"fmt"
	"github.com/jcmturner/restclient"
	"github.com/remmelt/evohome-prometheus-export/logging"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	applicationID = "b013aa26-9724-4dbd-8897-048b9aada249"
	//applicationID = "91db1612-73fd-4500-91b2-e63b069b185c"
	authUrl = "/Auth/OAuth/Token"
)

type idHeaders struct {
	Authorization string
	ApplicationID string
}

type Authenticate struct {
	Request         *restclient.Request
	IdentityHeaders *idHeaders
	authResponse
	validUntil time.Time
	loggers    *logging.Loggers
	postData   *url.Values
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func (a *Authenticate) NewRequest(cfg *restclient.Config, logs *logging.Loggers) error {
	a.loggers = logs
	data := url.Values{}
	data.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	data.Set("Cache-Control", "no-store no-cache")
	data.Set("Pragma", "no-cache")
	data.Set("grant_type", "password")
	data.Set("scope", "EMEA-V1-Basic EMEA-V1-Anonymous EMEA-V1-Get-Current-User-Account")
	data.Set("Username", os.Getenv("EVOHOME_USERNAME"))
	data.Set("Password", os.Getenv("EVOHOME_PASSWORD"))
	a.postData = &data

	o := restclient.NewPostOperation().WithPath(authUrl).WithBodyDataURLValues(data).WithResponseTarget(a)

	cfg.WithUserId(applicationID)
	cfg.WithPassword("test")

	req, err := restclient.BuildRequest(cfg, o)
	if err != nil {
		return errors.New(fmt.Sprintf("Error building ReST request to authenticate: %v", err))
	}
	a.Request = req
	a.loggers.Info.Println("New authentication request object configured")
	return nil
}

func (a *Authenticate) Process() error {
	if a.AccessToken == "" || time.Now().After(a.validUntil) {
		a.loggers.Info.Println("No OAuth token available or it has expired. Requesting one.")
		err := a.callAuthService()
		if err != nil {
			return err
		}
		id := idHeaders{
			Authorization: fmt.Sprintf("%s %s", a.TokenType, a.AccessToken),
			ApplicationID: applicationID,
		}
		a.IdentityHeaders = &id
	} else {
		a.loggers.Info.Println("OAuth token still valid.")
	}
	return nil
}

func (a *Authenticate) callAuthService() error {
	//Have to build the request again as the send data gets closed.
	req, err := restclient.BuildRequest(a.Request.Config, a.Request.Operation)
	if err != nil {
		return errors.New(fmt.Sprintf("Error building ReST request to authenticate: %v", err))
	}
	a.Request = req
	a.loggers.Info.Println("New authentication request object configured")
	code, e := restclient.Send(a.Request)
	if e != nil {
		return errors.New(fmt.Sprintf("Authentication error, HTTP code %v; %v", *code, e))
	}
	if *code != http.StatusOK {
		return errors.New(fmt.Sprintf("Authentication error, got HTTP status %v rather than HTTP status %v from authentication call to %v.", *code, http.StatusOK, a.Request.HTTPRequest.URL.String()))
	}
	if a.ExpiresIn > 0 {
		a.validUntil = time.Now().Add(time.Duration(a.ExpiresIn) * time.Second)
		a.loggers.Info.Printf("OAuth token retrieved. Valid until %v\n", a.validUntil)
	}
	return nil
}
