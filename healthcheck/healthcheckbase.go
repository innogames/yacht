package healthcheck

import (
	"github.com/innogames/yacht/logger"
	"strconv"
	"sync"
	"time"
)

// HealthCheck is a structure holding properties shared between all healthcheck types.
type HealthCheckBase struct {
	// Properties
	hcType         string
	interval       int
	timeout        int
	maxFailed      int
	minNodes       int
	minNodesAction string
	maxNodes       int

	// Communication
	stopChan chan bool

	// Implemented interface
	HealthCheck
}

// configure sets up base properties of a healthcheck with reasonable defaults.
func (this *HealthCheckBase) configure(json JSONMap) {
	this.stopChan = make(chan bool)

	// FIXME: fix it in lbadmin
	// Miminal interval is 1s.
	this.interval, _ = strconv.Atoi(json["interval"].(string))
	if this.interval < 1 {
		this.interval = 1
	}

	// Minimal timeout is 1s.
	if this.timeout < 1000 {
		this.timeout = 1000
	}

}

func (this *HealthCheckBase) run(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-this.stopChan:
			return
		case <-time.After(time.Second * time.Duration(this.interval)):
			logger.Info.Printf("this %v Running", this)
		}
	}
}

func (this *HealthCheckBase) Stop() {
	this.stopChan <- true
}
