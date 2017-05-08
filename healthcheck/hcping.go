package healthcheck

import (
	"github.com/innogames/yacht/logger"
	"sync"
)

// HCPing stores all properties of a ping healthcheck. In fact it stores nothing
// because all required parameters are stored in HCbase.
type HCPing struct {
	Base
}

// NewHCPing creates new ping healthcheck struct and populates it with data from Json config.
func newHCPing(json JSONMap) *HCPing {
	hc := new(HCPing)
	hc.Base.configure(json)
	hc.hcType = json["type"].(string)
	logger.Info.Printf("healthcheck: %s", hc.hcType)
	return hc
}

// run starts operation of this healthcheck, in fact it calls the Base class.
func (hc *HCPing) run(wg *sync.WaitGroup) {
	hc.Base.run(wg)
}

// Stop terminates this healthcheck, in fact it calls the Base class.
func (hc *HCPing) Stop() {
	hc.Base.Stop()
}
