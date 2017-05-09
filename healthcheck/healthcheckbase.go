package healthcheck

import (
	"context"
	"github.com/innogames/yacht/logger"
	"sync"
	"time"
)

// HCBase is a structure holding properties shared between all healthcheck types.
type HCBase struct {
	// Properties - configuration
	hcType         string
	ipAddress      string
	interval       int
	timeout        int
	maxFailed      int
	minNodes       int
	minNodesAction string
	maxNodes       int

	// Properties - operation
	failures  int
	prevState int

	// Communication
	logPrefix string
	stopChan  chan bool

	// Implemented interface
	HealthCheck
}

func jsonIntDefault(json JSONMap, key string, dflt int) int {
	if val, ok := json[key].(int); ok && val >= dflt {
		return val
	}
	return dflt
}

// configure sets up base properties of a healthcheck with reasonable defaults.
func (hcb *HCBase) configure(json JSONMap, ipAddress string) {
	// logPrefix is not configured here because it might be slightly different for each type of HealthCheck
	hcb.stopChan = make(chan bool)
	hcb.ipAddress = ipAddress

	// Read configuration parameters from JSON or provide a reasonable default.
	hcb.maxFailed = jsonIntDefault(json, "maxFailed", 3)
	hcb.interval = jsonIntDefault(json, "interval", 1)
	hcb.timeout = jsonIntDefault(json, "timeout", 1000)
}

// run starts operation of a healthcheck. It is an endless loop running in a goroutine
// which performs the real checking operation in scheduled time intervals. It can be
// terminated by sending a boolean over stopChan.
func (hcb *HCBase) run(wg *sync.WaitGroup, do func(chan Result) context.CancelFunc) {
	defer wg.Done()

	for {
		// Prepare a chanel to receive results from and do the Healthcheck.
		resChan := make(chan Result)
		cancel := do(resChan)

		// Wait for finish of do() or end of program.
		select {
		case res := <-resChan:
			lastState := res.ret
			if hcb.prevState != hcGood && lastState == hcGood {
				hcb.failures = 0
				logger.Info.Printf(hcb.logPrefix + " passed")
			}
			if lastState != hcGood && hcb.failures < hcb.maxFailed {
				hcb.failures++
				logger.Info.Printf(hcb.logPrefix+" failed %d/%d reason: %s", hcb.failures, hcb.maxFailed, res.err)
				if hcb.failures == hcb.maxFailed {
				}
			}
			hcb.prevState = lastState
		case <-hcb.stopChan:
			if cancel != nil {
				// Terminate already running check.
				cancel()
			}
			return
		}

		// Now additionaly to time it took to make the check wait for the specified interval.
		// End of program can be sent to us when waiting, so again catch it.
		select {
		case <-time.After(time.Second * time.Duration(hcb.interval)):
			if cancel != nil {
				cancel()
			}
		case <-hcb.stopChan:
			return
		}
	}
}

// Stop terminates this healthcheck. It works by sending a boolean over stopChan
// to the main goroutine of this check.
func (hcb *HCBase) Stop() {
	hcb.stopChan <- true
}
