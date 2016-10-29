package userAccount

import (
	"errors"
	"fmt"
	"github.com/jcmturner/evohome-prometheus-export/authenticate"
	"github.com/jcmturner/restclient"
	"net/http"
	"github.com/jcmturner/evohome-prometheus-export/logging"
)

const (
	url = "/WebAPI/emea/api/v1/userAccount"
)

type UserAccount struct {
	Request            *restclient.Request
	userAccountDetails
	loggers		*logging.Loggers
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

func (u *UserAccount) NewRequest(cfg *restclient.Config, logs *logging.Loggers) error {
	u.loggers = logs
	o := restclient.NewGetOperation().WithPath(url).WithResponseTarget(u)
	req, err := restclient.BuildRequest(cfg, o)
	if err != nil {
		return errors.New(fmt.Sprintf("Error building ReST request to authenticate: %v", err))
	}
	u.Request = req
	u.loggers.Info.Println("New userAccount request object configured")
	return nil
}

func (u *UserAccount) process(a *authenticate.Authenticate) error {
	// Details will not be refreshed. A restart would be needed.
	if u.UserID != "" {
		u.loggers.Info.Println("UserID information already available. Returning from cache. Restart required to refresh.")
		return nil
	}
	u.loggers.Info.Println("UserID information not available. Requesting...")
	err := a.Process()
	if err != nil {
		return err
	}
	u.Request.HTTPRequest.Header.Set("Authorization", a.IdentityHeaders.Authorization)
	u.Request.HTTPRequest.Header.Set("applicationId", a.IdentityHeaders.ApplicationID)
	code, e := restclient.Send(u.Request)
	if e != nil {
		return errors.New(fmt.Sprintf("UserAccount error, HTTP code %v; %v", *code, e))
	}
	if *code != http.StatusOK {
		return errors.New(fmt.Sprintf("UserAccount error, got HTTP status %v rather than HTTP status %v from authentication call to %v.", *code, http.StatusOK, u.Request.HTTPRequest.URL.String()))
	}
	return nil
}

func (u *UserAccount) GetUserID(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return "", err
	}
	return u.UserID, nil
}

func (u *UserAccount) GetCity(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return "", err
	}
	return u.City, nil
}

func (u *UserAccount) GetCountry(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return "", err
	}
	return u.Country, nil
}

func (u *UserAccount) GetFirstname(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return "", err
	}
	return u.Firstname, nil
}

func (u *UserAccount) GetLanguage(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return "", err
	}
	return u.Language, nil
}

func (u *UserAccount) GetLastname(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return "", err
	}
	return u.Lastname, nil
}

func (u *UserAccount) GetPostcode(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return "", err
	}
	return u.Postcode, nil
}

func (u *UserAccount) GetStreetAddress(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return "", err
	}
	return u.StreetAddress, nil
}

func (u *UserAccount) GetUsername(a *authenticate.Authenticate) (string, error) {
	err := u.process(a)
	if err != nil {
		return "", err
	}
	return u.Username, nil
}
