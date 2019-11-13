package handlers

import (
	"fmt"
	"net/http"

	"github.com/remmelt/evohome-prometheus-export/authenticate"
	"github.com/remmelt/evohome-prometheus-export/location"
	"github.com/remmelt/evohome-prometheus-export/logging"
)

// GetZoneTemperatures print zone temperature to prometheus format
func GetZoneTemperatures(w http.ResponseWriter, a *authenticate.Authenticate, l *location.Location, logs *logging.Loggers) {
	zones, err := l.GetTemperatureControlSystemZonesStatus(a)
	if err != nil {
		logs.Error.Printf("Could not get zone information: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	setNoCacheHeaders(w)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	for _, z := range zones {
		fmt.Fprintf(w, "evohome_current_temperature{label=%q} %v\n", z.Name, z.CurrentTemperature)
		fmt.Fprintf(w, "evohome_target_temperature{label=%q} %v\n", z.Name, z.TargetTemperature)
	}
	return
}

func setNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}
