package healthcheck

import (
	"github.com/innogames/yacht/logger"
	"strconv"
	"sync"
	"time"
)

// Base is a structure holding properties shared between all healthcheck types.
type Base struct {
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
func (hcb *Base) configure(json JSONMap) {
	hcb.stopChan = make(chan bool)

	// FIXME: fix it in lbadmin
	// Miminal interval is 1s.
	hcb.interval, _ = strconv.Atoi(json["interval"].(string))
	if hcb.interval < 1 {
		hcb.interval = 1
	}

	// Minimal timeout is 1s.
	if hcb.timeout < 1000 {
		hcb.timeout = 1000
	}

}

// run starts operation of a healthcheck. It is an endless loop running in a goroutine
// which performs the real checking operation in scheduled time intervals. It can be
// terminated by sending a boolean over stopChan.
func (hcb *Base) run(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-hcb.stopChan:
			return
		case <-time.After(time.Second * time.Duration(hcb.interval)):
			logger.Info.Printf("this %v Running", hcb)
		}
	}
}

// Stop terminates this healthcheck. It works by sending a boolean over stopChan
// to the main goroutine of this check.
func (hcb *Base) Stop() {
	hcb.stopChan <- true
}
