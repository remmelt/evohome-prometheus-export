package userAccount

import (
	"errors"
	"fmt"
	"github.com/jcmturner/evohome-prometheus-export/authenticate"
	"github.com/jcmturner/restclient"
	"net/http"
)

const (
	url = "https://tccna.honeywell.com/WebAPI/emea/api/v1/userAccount"
)

type UserAccount struct {
	Request            *restclient.Request
	UserAccountDetails *userAccountDetails
}

type userAccountDetails struct {
	UserID        string `json:"userId"`
	Username      string `json:"username"`
	Firstname     string `json:"firstname"`
	Lastname      string `json:"lastname"`
	StreetAddress string `json:"streetAddress"`
	City          string `json:"city"`
	Postcode      string `json:"postcode"`
	Country       string `json:"country"`
	Language      string `json:"language"`
}

func (u *UserAccount) NewRequest() error {
	o := restclient.NewPostOperation()
	o.WithPath(url)
	cfg := restclient.NewConfig()
	req, err := restclient.BuildRequest(cfg, o)
	if err != nil {
		return errors.New(fmt.Sprintf("Error building ReST request to authenticate: %v", err))
	}
	u.Request = req
	return nil
}

func (u *UserAccount) process(a *authenticate.Authenticate) error {
	// Details will not be refreshed. A restart would be needed.
	if u.UserAccountDetails != nil {
		return nil
	}
	a.Process()
	u.Request.HTTPRequest.Header.Set("Authorization", a.IdentityHeaders.Authorization)
	u.Request.HTTPRequest.Header.Set("applicationId", a.IdentityHeaders.ApplicationID)
	a.Request.Operation.WithResponseTarget(&u.UserAccountDetails)
	code, e := restclient.Send(u.Request)
	if e != nil {
		return e
	}
	if *code != http.StatusOK {
		return errors.New(fmt.Sprintf("UserAccount error, got HTTP status %v rather than HTTP status %v from authentication call to %v.", *code, http.StatusOK, a.Request.HTTPRequest.URL.String()))
	}
	return nil
}

func (u *UserAccount) GetUserID(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return nil, err
	}
	return u.UserAccountDetails.UserID
}

func (u *UserAccount) GetCity(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return nil, err
	}
	return u.UserAccountDetails.City
}

func (u *UserAccount) GetCountry(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return nil, err
	}
	return u.UserAccountDetails.Country
}

func (u *UserAccount) GetFirstname(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return nil, err
	}
	return u.UserAccountDetails.Firstname
}

func (u *UserAccount) GetLanguage(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return nil, err
	}
	return u.UserAccountDetails.Language
}

func (u *UserAccount) GetLastname(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return nil, err
	}
	return u.UserAccountDetails.Lastname
}

func (u *UserAccount) GetPostcode(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return nil, err
	}
	return u.UserAccountDetails.Postcode
}

func (u *UserAccount) GetStreetAddress(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return nil, err
	}
	return u.UserAccountDetails.StreetAddress
}

func (u *UserAccount) GetUsername(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return nil, err
	}
	return u.UserAccountDetails.Username
}
