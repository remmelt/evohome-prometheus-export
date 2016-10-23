package location

import (
	"github.com/jcmturner/restclient"
	"errors"
	"fmt"
	"net/http"
	"github.com/jcmturner/evohome-prometheus-export/authenticate"
)

const (
	url = "https://tccna.honeywell.com/WebAPI/emea/api/v1/location"
)

type Location struct {
	Request         *restclient.Request
	LocationStatus *LocationStatus
}

type ZoneStatus struct {
	Name string
	ZoneID string
	CurrentTemperature float32
	TargetTemperature float32
	SetpointMode string
}

type LocationStatus struct {
	LocationID string `json:"locationId"`
	Gateways []struct {
		GatewayID string `json:"gatewayId"`
		TemperatureControlSystems []struct {
			SystemID string `json:"systemId"`
			Zones []struct {
				ZoneID string `json:"zoneId"`
				TemperatureStatus struct {
					       Temperature float32 `json:"temperature"`
					       IsAvailable bool `json:"isAvailable"`
				       } `json:"temperatureStatus"`
				ActiveFaults []interface{} `json:"activeFaults"`
				HeatSetpointStatus struct {
					       TargetTemperature float32 `json:"targetTemperature"`
					       SetpointMode string `json:"setpointMode"`
				       } `json:"heatSetpointStatus"`
				Name string `json:"name"`
			} `json:"zones"`
			ActiveFaults []interface{} `json:"activeFaults"`
			SystemModeStatus struct {
					 Mode string `json:"mode"`
					 IsPermanent bool `json:"isPermanent"`
				 } `json:"systemModeStatus"`
		} `json:"temperatureControlSystems"`
		ActiveFaults []interface{} `json:"activeFaults"`
	} `json:"gateways"`
}

func (l *Location) NewRequest(id string) error {
	o := restclient.NewPostOperation()
	o.WithPath(fmt.Sprintf("%v/%v/status?includeTemperatureControlSystems=True", url, id))
	cfg := restclient.NewConfig()
	req, err := restclient.BuildRequest(cfg, o)
	if err != nil {
		return errors.New(fmt.Sprintf("Error building ReST request to authenticate: %v", err))
	}
	l.Request = req
	return nil
}

func (l *Location) process(a *authenticate.Authenticate) error {
	// Details will not be refreshed. A restart would be needed.
	if l.LocationStatus != nil {
		return nil
	}
	a.Process()
	l.Request.HTTPRequest.Header.Set("Authorization", a.IdentityHeaders.Authorization)
	l.Request.HTTPRequest.Header.Set("applicationId", a.IdentityHeaders.ApplicationID)
	a.Request.Operation.WithResponseTarget(&l.LocationStatus)
	code, e := restclient.Send(l.Request)
	if e != nil {
		return e
	}
	if *code != http.StatusOK {
		return errors.New(fmt.Sprintf("Location error, got HTTP status %v rather than HTTP status %v from authentication call to %v.", *code, http.StatusOK, a.Request.HTTPRequest.URL.String()))
	}
	return nil
}

func (l *Location) GetTemperatureControlSystemZonesStatus(a *authenticate.Authenticate) ([]ZoneStatus, error) {
	err := l.process(a)
	if err != nil {
		return nil, err
	}
	zones:= make([]ZoneStatus, len(l.LocationStatus.Gateways[0].TemperatureControlSystems[0].Zones))
	for i, z := range l.LocationStatus.Gateways[0].TemperatureControlSystems[0].Zones {
		zones[i] = ZoneStatus{
			Name: z.Name,
			ZoneID: z.ZoneID,
			CurrentTemperature: z.TemperatureStatus.Temperature,
			TargetTemperature: z.HeatSetpointStatus.TargetTemperature,
			SetpointMode: z.HeatSetpointStatus.SetpointMode,
		}
	}
	return zones, nil
}
