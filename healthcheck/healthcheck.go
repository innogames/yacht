package healthcheck

import (
	"github.com/innogames/yacht/logger"
	"sync"
)

// JSONMap is a shortcut for a JSON dictionary.
type JSONMap map[string]interface{}

// HealthCheck defines which functions must every type of healthcheck implement.
type HealthCheck interface {
	Run(wg *sync.WaitGroup)
	Stop()
}

// NewHealthCheck is an object factory returning a proper HealtCheck object depending
// in configuration it reads from JSON and starts its main goroutine.
func NewHealthCheck(lbNodeChan chan bool, logPrefix string, json JSONMap, ipAddress string) *HealthCheck {
	hctype := json["type"].(string)

	var hc HealthCheck
	var hcb *HCBase

	switch hctype {
	case "dummy":
		hc, hcb = newHCDummy(logPrefix, json)
	case "http":
		hc, hcb = newHCHttp(logPrefix, json)
	case "https":
		hc, hcb = newHCHttp(logPrefix, json)
	case "ping":
		hc, hcb = newHCPing(json)
	case "script":
		hc, hcb = newHCScript(json)
	default:
		logger.Error.Printf(logPrefix+"Unknown HealthCheck type %s", hctype)
		return nil
	}
	hcb.configure(lbNodeChan, json, ipAddress)

	return &hc
}
