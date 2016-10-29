package installation

import (
	"errors"
	"fmt"
	"github.com/jcmturner/evohome-prometheus-export/authenticate"
	"github.com/jcmturner/restclient"
	"net/http"
	"net/url"
	"github.com/jcmturner/evohome-prometheus-export/logging"
)

const (
	instUrl = "/WebAPI/emea/api/v1/location/installationInfo"
)

type Installation struct {
	Request          *restclient.Request
	InstallationInfo *[]installationInfo
	loggers		*logging.Loggers
}

type ZoneInfo struct {
	Name   string
	ZoneID string
}

type installationInfo struct {
	LocationInfo struct {
		LocationID               string `json:"locationId"`
		Name                     string `json:"name"`
		StreetAddress            string `json:"streetAddress"`
		City                     string `json:"city"`
		Country                  string `json:"country"`
		Postcode                 string `json:"postcode"`
		LocationType             string `json:"locationType"`
		UseDaylightSaveSwitching bool   `json:"useDaylightSaveSwitching"`
		TimeZone                 struct {
			TimeZoneID             string `json:"timeZoneId"`
			DisplayName            string `json:"displayName"`
			OffsetMinutes          int    `json:"offsetMinutes"`
			CurrentOffsetMinutes   int    `json:"currentOffsetMinutes"`
			SupportsDaylightSaving bool   `json:"supportsDaylightSaving"`
		} `json:"timeZone"`
		LocationOwner struct {
			UserID    string `json:"userId"`
			Username  string `json:"username"`
			Firstname string `json:"firstname"`
			Lastname  string `json:"lastname"`
		} `json:"locationOwner"`
	} `json:"locationInfo"`
	Gateways []struct {
		GatewayInfo struct {
			GatewayID string `json:"gatewayId"`
			Mac       string `json:"mac"`
			Crc       string `json:"crc"`
			IsWiFi    bool   `json:"isWiFi"`
		} `json:"gatewayInfo"`
		TemperatureControlSystems []struct {
			SystemID  string `json:"systemId"`
			ModelType string `json:"modelType"`
			Zones     []struct {
				ZoneID                   string `json:"zoneId"`
				ModelType                string `json:"modelType"`
				HeatSetpointCapabilities struct {
					MaxHeatSetpoint      float32      `json:"maxHeatSetpoint"`
					MinHeatSetpoint      float32      `json:"minHeatSetpoint"`
					ValueResolution      float32  `json:"valueResolution"`
					AllowedSetpointModes []string `json:"allowedSetpointModes"`
					MaxDuration          string   `json:"maxDuration"`
					TimingResolution     string   `json:"timingResolution"`
				} `json:"heatSetpointCapabilities"`
				ScheduleCapabilities struct {
					MaxSwitchpointsPerDay   int     `json:"maxSwitchpointsPerDay"`
					MinSwitchpointsPerDay   int     `json:"minSwitchpointsPerDay"`
					TimingResolution        string  `json:"timingResolution"`
					SetpointValueResolution float32 `json:"setpointValueResolution"`
				} `json:"scheduleCapabilities"`
				Name     string `json:"name"`
				ZoneType string `json:"zoneType"`
			} `json:"zones"`
			AllowedSystemModes []struct {
				SystemMode       string `json:"systemMode"`
				CanBePermanent   bool   `json:"canBePermanent"`
				CanBeTemporary   bool   `json:"canBeTemporary"`
				MaxDuration      string `json:"maxDuration,omitempty"`
				TimingResolution string `json:"timingResolution,omitempty"`
				TimingMode       string `json:"timingMode,omitempty"`
			} `json:"allowedSystemModes"`
		} `json:"temperatureControlSystems"`
	} `json:"gateways"`
}

func (i *Installation) NewRequest(userID string, cfg *restclient.Config, logs *logging.Loggers) error {
	i.loggers = logs
	var iInfo []installationInfo
	i.InstallationInfo = &iInfo
	o := restclient.NewGetOperation().WithPath(instUrl).WithResponseTarget(i.InstallationInfo)
	data := url.Values{}
	data.Set("includeTemperatureControlSystems", "True")
	data.Set("userId", userID)
	o.WithQueryDataURLValues(data)
	req, err := restclient.BuildRequest(cfg, o)
	if err != nil {
		return errors.New(fmt.Sprintf("Error building ReST request to authenticate: %v", err))
	}
	i.Request = req
	i.loggers.Info.Println("New installation request object configured")
	return nil
}

func (i *Installation) process(a *authenticate.Authenticate) error {
	// Details will not be refreshed. A restart would be needed.
	if len(*i.InstallationInfo) > 0 {
		i.loggers.Info.Println("Intallation information already available. Returning from cache. Restart required to refresh.")
		return nil
	}
	i.loggers.Info.Println("Installation information not available. Requesting...")
	err := a.Process()
	if err != nil {
		return err
	}
	i.Request.HTTPRequest.Header.Set("Authorization", a.IdentityHeaders.Authorization)
	i.Request.HTTPRequest.Header.Set("applicationId", a.IdentityHeaders.ApplicationID)
	code, e := restclient.Send(i.Request)
	if e != nil {
		return errors.New(fmt.Sprintf("Installation error, HTTP code %v; %v", *code, e))
	}
	if *code != http.StatusOK {
		return errors.New(fmt.Sprintf("Installation error, got HTTP status %v rather than HTTP status %v from authentication call to %v.", *code, http.StatusOK, i.Request.HTTPRequest.URL.String()))
	}
	return nil
}

func (i *Installation) GetLocationID(a *authenticate.Authenticate) (string, error) {
	err := i.process(a)
	if err != nil  {
		return "", err
	}
	if len(*i.InstallationInfo) < 1 {
		return "", errors.New("Did not get any installations in the response.")
	}
	return (*i.InstallationInfo)[0].LocationInfo.LocationID, nil
}

func (i *Installation) GetSystemID(a *authenticate.Authenticate) (string, error) {
	err := i.process(a)
	if err != nil {
		return "", err
	}
	if len(*i.InstallationInfo) < 1 {
		return "", errors.New("Did not get any installations in the response.")
	}
	return (*i.InstallationInfo)[0].Gateways[0].TemperatureControlSystems[0].SystemID, nil
}

func (i *Installation) GetTemperatureControlSystemZones(a *authenticate.Authenticate) ([]ZoneInfo, error) {
	err := i.process(a)
	if err != nil {
		return nil, err
	}
	if len(*i.InstallationInfo) < 1 {
		return nil, errors.New("Did not get any installations in the response.")
	}
	zones := make([]ZoneInfo, len((*i.InstallationInfo)[0].Gateways[0].TemperatureControlSystems[0].Zones))
	for i, z := range (*i.InstallationInfo)[0].Gateways[0].TemperatureControlSystems[0].Zones {
		zones[i] = ZoneInfo{Name: z.Name, ZoneID: z.ZoneID}
	}
	return zones, nil
}
