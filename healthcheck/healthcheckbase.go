package healthcheck

import (
	"github.com/innogames/yacht/logger"
	"net"
	"sync"
	"time"
)

// HCBase is a structure holding properties shared between all healthcheck types.
type HCBase struct {
	// Implemented interface
	HealthCheck

	// Configuration
	hcType         string
	ipAddress      net.IP
	interval       int
	timeout        int
	maxFailed      int
	minNodes       int
	minNodesAction string
	maxNodes       int

	// Operation
	failures  int
	prevState HCResult

	// Communication
	logPrefix  string
	lbNodeChan chan HCResultMsg
	stopChan   chan bool
}

func jsonIntDefault(json JSONMap, key string, dflt int) int {
	if val, ok := json[key].(float64); ok && int(val) >= dflt {
		return int(val)
	}
	return dflt
}

// configure sets up base properties of a healthcheck with reasonable defaults.
func (hcb *HCBase) configure(lbNodeChan chan HCResultMsg, json JSONMap, ipAddress string) {
	// logPrefix is not configured here because it might be slightly different for each type of HealthCheck
	hcb.stopChan = make(chan bool)
	hcb.lbNodeChan = lbNodeChan
	hcb.ipAddress = net.ParseIP(ipAddress)

	// Read configuration parameters from JSON or provide a reasonable default.
	hcb.maxFailed = jsonIntDefault(json, "maxFailed", 3)
	hcb.interval = jsonIntDefault(json, "interval", 1)
	hcb.timeout = jsonIntDefault(json, "timeout", 1000)
}

// run starts operation of a healthcheck. It is an endless loop running in a goroutine
// which performs the real checking operation in scheduled time intervals. It can be
// terminated by sending a boolean over stopChan.
func (hcb *HCBase) run(wg *sync.WaitGroup, hc HealthCheck) {
	wg.Add(1)
	defer wg.Done()

	for {
		// Prepare a chanel to receive results from and do the Healthcheck.
		resChan := make(chan HCResultError)
		cancel := hc.do(resChan)

		// Wait for finish of do() or end of program.
		select {
		case res := <-resChan:
			lastState := res.res
			if hcb.prevState != HCGood && lastState == HCGood {
				hcb.failures = 0
				logger.Info.Printf(hcb.logPrefix + "action: passed")
				hcb.lbNodeChan <- HCResultMsg{
					result: HCGood,
					HC:     hc,
				}
			}
			if lastState != HCGood && hcb.failures < hcb.maxFailed {
				hcb.failures++
				logger.Info.Printf(hcb.logPrefix+"action: failed %d/%d reason: %s", hcb.failures, hcb.maxFailed, res.err)
				if hcb.failures == hcb.maxFailed {
					hcb.lbNodeChan <- HCResultMsg{
						result: HCBad,
						HC:     hc,
					}
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
