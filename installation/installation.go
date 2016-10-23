package installation

import (
	"github.com/jcmturner/restclient"
	"errors"
	"fmt"
	"net/http"
	"github.com/jcmturner/evohome-prometheus-export/authenticate"
)

const (
	url = "https://tccna.honeywell.com/WebAPI/emea/api/v1/location/installationInfo?includeTemperatureControlSystems=True&userId="
)

type Location struct {
	Request         *restclient.Request
	InstallationInfo *installationInfo
}

type ZoneInfo struct {
	Name string
	ZoneID string
}

type installationInfo []struct {
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
								 MaxHeatSetpoint      int      `json:"maxHeatSetpoint"`
								 MinHeatSetpoint      int      `json:"minHeatSetpoint"`
								 ValueResolution      float64  `json:"valueResolution"`
								 AllowedSetpointModes []string `json:"allowedSetpointModes"`
								 MaxDuration          string   `json:"maxDuration"`
								 TimingResolution     string   `json:"timingResolution"`
							 } `json:"heatSetpointCapabilities"`
				ScheduleCapabilities struct {
								 MaxSwitchpointsPerDay   int     `json:"maxSwitchpointsPerDay"`
								 MinSwitchpointsPerDay   int     `json:"minSwitchpointsPerDay"`
								 TimingResolution        string  `json:"timingResolution"`
								 SetpointValueResolution float64 `json:"setpointValueResolution"`
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

func (l *Location) NewRequest(userID string) error {
	o := restclient.NewPostOperation()
	o.WithPath(fmt.Sprintf("%v%v", url, userID))
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
	if l.InstallationInfo != nil {
		return nil
	}
	a.Process()
	l.Request.HTTPRequest.Header.Set("Authorization", a.IdentityHeaders.Authorization)
	l.Request.HTTPRequest.Header.Set("applicationId", a.IdentityHeaders.ApplicationID)
	a.Request.Operation.WithResponseTarget(&l.InstallationInfo)
	code, e := restclient.Send(l.Request)
	if e != nil {
		return e
	}
	if *code != http.StatusOK {
		return errors.New(fmt.Sprintf("Location error, got HTTP status %v rather than HTTP status %v from authentication call to %v.", *code, http.StatusOK, a.Request.HTTPRequest.URL.String()))
	}
	return nil
}

func (l *Location) GetLocationID(a *authenticate.Authenticate) (string, error) {
	err := l.process(a)
	if err != nil {
		return nil, err
	}
	return l.InstallationInfo[0].LocationInfo.LocationID, nil
}

func (l *Location) GetSystemID(a *authenticate.Authenticate) (string, error) {
	err := l.process(a)
	if err != nil {
		return nil, err
	}
	return l.InstallationInfo[0].Gateways[0].TemperatureControlSystems[0].SystemID, nil
}

func (l *Location) GetTemperatureControlSystemZones(a *authenticate.Authenticate) ([]ZoneInfo, error) {
	err := l.process(a)
	if err != nil {
		return nil, err
	}
	zones:= make([]ZoneInfo, len(l.InstallationInfo[0].Gateways[0].TemperatureControlSystems[0].Zones))
	for i, z := range l.InstallationInfo[0].Gateways[0].TemperatureControlSystems[0].Zones {
		zones[i] = ZoneInfo{Name: z.Name, ZoneID: z.ZoneID}
	}
	return zones, nil
}
