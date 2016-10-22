package authenticate

import (
	"errors"
	"fmt"
	"github.com/jcmturner/restclient"
	"net/http"
	"os"
	"time"
)

const (
	applicationID = "b013aa26-9724-4dbd-8897-048b9aada249"
	//applicationID = "91db1612-73fd-4500-91b2-e63b069b185c"
	authUrl = "https://tccna.honeywell.com/Auth/OAuth/Token"
)

type authStruct struct {
	ContentType  string `json:"Content-Type"`
	Host         string `json:"Host"`
	CacheControl string `json:"Cache-Control"`
	Pragma       string `json:"Pragma"`
	GrantType    string `json:"grant_type"`
	Scope        string `json:"scope"`
	Username     string `json:"Username"`
	Password     string `json:"Password"`
	Connection   string `json:"Connection"`
}

type idHeaders struct {
	Authorization string
	ApplicationID string
}

type Authenticate struct {
	Request         *restclient.Request
	IdentityHeaders *idHeaders
	authResponse    *authResponse
	validUntil      time.Time
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func (a *Authenticate) NewRequest() error {
	data := authStruct{
		ContentType:  "application/x-www-form-urlencoded; charset = utf-8",
		CacheControl: "no-store no-cache",
		Pragma:       "no-cache",
		GrantType:    "password",
		Scope:        "EMEA-V1-Basic EMEA-V1-Anonymous EMEA-V1-Get-Current-User-Account",
		Connection:   "Keep-Alive",
		Username:     os.Getenv("EVOHOME_USERNAME"),
		Password:     os.Getenv("EVOHOME_PASSWORD"),
	}
	o := restclient.NewPostOperation()
	o.WithPath(authUrl)
	o.WithBodyDataStruct(data)
	cfg := restclient.NewConfig()
	cfg.WithUserId(applicationID)
	cfg.WithPassword("test")
	req, err := restclient.BuildRequest(cfg, o)
	if err != nil {
		return errors.New(fmt.Sprintf("Error building ReST request to authenticate: %v", err))
	}
	a.Request = req
	return nil
}

func (a *Authenticate) Process() (*idHeaders, error) {
	if a.authResponse == nil {
		err := a.callAuthService()
		if err != nil {
			return nil, err
		}
	} else {
		if !a.validUntil.IsZero() && time.Now().After(a.validUntil) {
			err := a.callAuthService()
			if err != nil {
				return nil, err
			}
		}
	}
	id := idHeaders{
		Authorization: "bearer " + a.authResponse.AccessToken,
		ApplicationID: applicationID,
	}
	a.IdentityHeaders = &id
	return a.IdentityHeaders, nil
}

func (a *Authenticate) callAuthService() error {
	r := authResponse{}
	a.Request.Operation.WithResponseTarget(&r)
	code, e := restclient.Send(a.Request)
	if e != nil {
		return e
	}
	if *code != http.StatusOK {
		return errors.New(fmt.Sprintf("Authentication error, got HTTP status %v rather than HTTP status %v from authentication call to %v.", *code, http.StatusOK, a.Request.HTTPRequest.URL.String()))
	}
	a.authResponse = &r
	if r.ExpiresIn > 0 {
		a.validUntil = time.Now().Add(time.Duration(r.ExpiresIn) * time.Second)
	}
	return nil
}
