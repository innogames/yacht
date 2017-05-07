package healthcheck

import (
	"github.com/innogames/yacht/logger"
	"sync"
	"time"
)

// JSONMap is a shortcut for a JSON dictionary.
type JSONMap map[string]interface{}

// HealthCheck defines which functions must every type of healthcheck implement.
type HealthCheck interface {
	Run(wg *sync.WaitGroup)
	Stop()
}

// HealthCheck is a structure holding properties shared between all healthcheck types.
type HealthCheckBase struct {
	hcType         string
	interval       int
	maxFailed      int
	minNodes       int
	minNodesAction string
	maxNodes       int

	stopChan chan bool
	HealthCheck
}

func NewHealthCheck(json JSONMap) *HealthCheck {
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

	return &hc
}

func (this HealthCheckBase) Run(wg *sync.WaitGroup) {
	// Increase counter of running Healtthishecks
	this.stopChan = make(chan bool)

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-this.stopChan:
				return
			case <-time.After(time.Second * time.Duration(this.interval)):
				logger.Info.Printf("this %v Running", this)
			}
		}
	}()
}

func (this HealthCheckBase) Stop() {
	this.stopChan <- true
}
