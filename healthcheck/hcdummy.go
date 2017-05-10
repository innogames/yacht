package healthcheck

import (
	"context"
	"fmt"
	"github.com/innogames/yacht/logger"
	"sync"
)

// HCDummy stores all properties of a dummy healthcheck.
type HCDummy struct {
	result Result
	HCBase
}

// NewHCDummy creates new ping healthcheck struct and populates it with data from Json config.
func newHCDummy(logPrefix string, json JSONMap) (*HCDummy, *HCBase) {
	hc := new(HCDummy)
	hc.hcType = json["type"].(string)

	if result, ok := json["result"].(Result); ok {
		hc.result = result
	}

	hc.logPrefix = logPrefix + fmt.Sprintf("healthcheck: %s response: %s ", hc.hcType, hc.result)

	logger.Info.Printf(hc.logPrefix + "created")
	return hc, &hc.HCBase
}

// do performs the healthckeck. It is called from the main goroutine of HealthcheckBase.
func (hc *HCDummy) do(hcr chan (ResultError)) context.CancelFunc {

	go func() {
		hcr <- ResultError{
			res: hc.result,
			err: nil,
		}
	}()
	return nil
}

// Run starts operation of this healthcheck, in fact it calls the Base class.
func (hc *HCDummy) Run(wg *sync.WaitGroup) {
	hc.HCBase.run(wg, hc.do)
}

// Stop terminates this healthcheck, in fact it calls the Base class.
func (hc *HCDummy) Stop() {
	hc.HCBase.Stop()
}
