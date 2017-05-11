package healthcheck

import (
	"context"
	"github.com/innogames/yacht/logger"
	"net"
	"sync"
)

// JSONMap is a shortcut for a JSON dictionary.
type JSONMap map[string]interface{}

// HealthCheck defines which functions must every type of healthcheck implement.
type HealthCheck interface {
	Run(wg *sync.WaitGroup)
	Stop()
	configure(lbNodeChan chan HCResultMsg, json JSONMap, ipAddress net.IP)
	do(hcr chan (HCResultError)) context.CancelFunc
}

// NewHealthCheck is an object factory returning a proper HealtCheck object depending
// in configuration it reads from JSON and starts its main goroutine.
func NewHealthCheck(lbNodeChan chan HCResultMsg, logPrefix string, json JSONMap, ipAddress net.IP) HealthCheck {
	hctype := json["type"].(string)

	var hc HealthCheck

	switch hctype {
	case "dummy":
		hc = newHCDummy(logPrefix, json)
	case "http":
		hc = newHCHttp(logPrefix, json)
	case "https":
		hc = newHCHttp(logPrefix, json)
	case "ping":
		hc = newHCPing(json)
	case "script":
		hc = newHCScript(json)
	default:
		logger.Error.Printf(logPrefix+"Unknown HealthCheck type %s", hctype)
		return nil
	}
	hc.configure(lbNodeChan, json, ipAddress)

	return hc
}
