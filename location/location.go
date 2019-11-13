package location

import (
	"errors"
	"fmt"
	"github.com/jcmturner/restclient"
	"github.com/remmelt/evohome-prometheus-export/authenticate"
	"github.com/remmelt/evohome-prometheus-export/logging"
	"net/http"
	"net/url"
)

const (
	apiurl = "/WebAPI/emea/api/v1/location"
)

type Location struct {
	Request *restclient.Request
	locationStatus
	loggers *logging.Loggers
}

type ZoneStatus struct {
	Name               string
	ZoneID             string
	CurrentTemperature float32
	TargetTemperature  float32
	SetpointMode       string
}

type locationStatus struct {
	LocationID string `json:"locationId"`
	Gateways   []struct {
		GatewayID                 string `json:"gatewayId"`
		TemperatureControlSystems []struct {
			SystemID string `json:"systemId"`
			Zones    []struct {
				ZoneID            string `json:"zoneId"`
				TemperatureStatus struct {
					Temperature float32 `json:"temperature"`
					IsAvailable bool    `json:"isAvailable"`
				} `json:"temperatureStatus"`
				ActiveFaults       []interface{} `json:"activeFaults"`
				HeatSetpointStatus struct {
					TargetTemperature float32 `json:"targetTemperature"`
					SetpointMode      string  `json:"setpointMode"`
				} `json:"heatSetpointStatus"`
				Name string `json:"name"`
			} `json:"zones"`
			ActiveFaults     []interface{} `json:"activeFaults"`
			SystemModeStatus struct {
				Mode        string `json:"mode"`
				IsPermanent bool   `json:"isPermanent"`
			} `json:"systemModeStatus"`
		} `json:"temperatureControlSystems"`
		ActiveFaults []interface{} `json:"activeFaults"`
	} `json:"gateways"`
}

func (l *Location) NewRequest(id string, cfg *restclient.Config, logs *logging.Loggers) error {
	l.loggers = logs
	data := url.Values{}
	data.Set("includeTemperatureControlSystems", "True")
	o := restclient.NewGetOperation().WithQueryDataURLValues(data).WithPath(fmt.Sprintf("%v/%v/status", apiurl, id))

	req, err := restclient.BuildRequest(cfg, o)
	if err != nil {
		return errors.New(fmt.Sprintf("Error building ReST request to authenticate: %v", err))
	}
	l.Request = req
	l.loggers.Info.Println("New location request object configured")
	return nil
}

func (l *Location) process(a *authenticate.Authenticate) error {
	l.loggers.Info.Println("Requesting latest location and zone information.")
	err := a.Process()
	if err != nil {
		return err
	}
	l.Request.HTTPRequest.Header.Set("Authorization", a.IdentityHeaders.Authorization)
	l.Request.HTTPRequest.Header.Set("applicationId", a.IdentityHeaders.ApplicationID)
	l.Request.Operation.WithResponseTarget(l)
	code, e := restclient.Send(l.Request)
	if e != nil {
		return errors.New(fmt.Sprintf("Location error, HTTP code %v; %v", *code, e))
	}
	if *code != http.StatusOK {
		return errors.New(fmt.Sprintf("Location error, got HTTP status %v rather than HTTP status %v from authentication call to %v.", *code, http.StatusOK, l.Request.HTTPRequest.URL.String()))
	}
	return nil
}

func (l *Location) GetTemperatureControlSystemZonesStatus(a *authenticate.Authenticate) ([]ZoneStatus, error) {
	err := l.process(a)
	if err != nil {
		return nil, err
	}
	zones := make([]ZoneStatus, len(l.Gateways[0].TemperatureControlSystems[0].Zones))
	for i, z := range l.Gateways[0].TemperatureControlSystems[0].Zones {
		zones[i] = ZoneStatus{
			Name:               z.Name,
			ZoneID:             z.ZoneID,
			CurrentTemperature: z.TemperatureStatus.Temperature,
			TargetTemperature:  z.HeatSetpointStatus.TargetTemperature,
			SetpointMode:       z.HeatSetpointStatus.SetpointMode,
		}
	}
	return zones, nil
}
