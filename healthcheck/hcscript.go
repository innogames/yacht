package healthcheck

import (
	"github.com/innogames/yacht/logger"
	"sync"
)

// HCScript stores all properties of a script healthcheck.
type HCScript struct {
	*HCBase
	Script string
}

// NewHCScript creates new script healthcheck struct and populates it with data from Json config.
func newHCScript(json JSONMap) *HCScript {
	hc := new(HCScript)
	hc.hcType = json["type"].(string)
	hc.Script = json["script"].(string)
	logger.Info.Printf("healthcheck: %s", hc.hcType)
	return hc
}

// Run starts operation of this healthcheck, in fact it calls the Base class.
func (hc *HCScript) Run(wg *sync.WaitGroup) {
	hc.HCBase.run(wg, hc)
}

// Stop terminates this healthcheck, in fact it calls the Base class.
func (hc *HCScript) Stop() {
	hc.HCBase.Stop()
}
