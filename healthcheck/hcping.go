package healthcheck

import (
	"github.com/innogames/yacht/logger"
	"sync"
)

// HCPing stores all properties of a ping healthcheck. In fact it stores nothing
// because all required parameters are stored in HCbase.
type HCPing struct {
	HCBase
}

// NewHCPing creates new ping healthcheck struct and populates it with data from Json config.
func newHCPing(json JSONMap) (*HCPing, *HCBase) {
	hc := new(HCPing)
	hc.hcType = json["type"].(string)
	logger.Info.Printf("healthcheck: %s", hc.hcType)
	return hc, &hc.HCBase
}

// Run starts operation of this healthcheck, in fact it calls the Base class.
func (hc *HCPing) Run(wg *sync.WaitGroup) {
	hc.HCBase.run(wg, nil)
}

// Stop terminates this healthcheck, in fact it calls the Base class.
func (hc *HCPing) Stop() {
	hc.HCBase.Stop()
}
