package healthcheck

import (
	"github.com/innogames/yacht/logger"
	"sync"
)

// JSONMap is a shortcut for a JSON dictionary.
type JSONMap map[string]interface{}

// Result is sent in a channel from a children, specific HC class to this master class.
type Result struct {
	ret int
	err error
}

const (
	hcError = iota
	hcBad
	hcGood
)

// HealthCheck defines which functions must every type of healthcheck implement.
type HealthCheck interface {
	run(wg *sync.WaitGroup)
	Stop()
}

// NewHealthCheck is an object factory returning a proper HealtCheck object depending
// in configuration it reads from JSON and starts its main goroutine.
func NewHealthCheck(wg *sync.WaitGroup, logPrefix string, json JSONMap, ipAddress string) *HealthCheck {
	hctype := json["type"].(string)

	var hc HealthCheck
	var hcb *HCBase

	switch hctype {
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
	hcb.configure(json, ipAddress)

	wg.Add(1) // Increase counter of running Healtthishecks
	go hc.run(wg)

	return &hc
}
