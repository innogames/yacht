package healthcheck

import (
	"github.com/innogames/yacht/logger"
	"sync"
)

// JSONMap is a shortcut for a JSON dictionary.
type JSONMap map[string]interface{}

// HealthCheck defines which functions must every type of healthcheck implement.
type HealthCheck interface {
	run(wg *sync.WaitGroup)
	Stop()
}

func NewHealthCheck(wg *sync.WaitGroup, json JSONMap) *HealthCheck {
	hctype := json["type"].(string)

	var hc HealthCheck

	switch hctype {
	case "http":
		hc = newHCHttp(json)
	case "https":
		hc = newHCHttp(json)
	case "ping":
		hc = newHCPing(json)
	case "script":
		hc = newHCScript(json)
	default:
		logger.Error.Printf("Unknown HealthCheck type %s", hctype)
		return nil
	}

	wg.Add(1) // Increase counter of running Healtthishecks
	go hc.run(wg)

	return &hc
}
