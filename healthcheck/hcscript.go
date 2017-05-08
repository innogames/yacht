package healthcheck

import (
	"github.com/innogames/yacht/logger"
	"sync"
)

// HCScript stores all properties of a script healthcheck.
type HCScript struct {
	Base
	Script string
}

// NewHCScript creates new script healthcheck struct and populates it with data from Json config.
func newHCScript(json JSONMap) *HCScript {
	hc := new(HCScript)
	hc.Base.configure(json)
	hc.hcType = json["type"].(string)
	hc.Script = json["script"].(string)
	logger.Info.Printf("healthcheck: %s", hc.hcType)
	return hc
}

// run starts operation of this healthcheck, in fact it calls the Base class.
func (hc *HCScript) run(wg *sync.WaitGroup) {
	hc.Base.run(wg)
}

// Stop terminates this healthcheck, in fact it calls the Base class.
func (hc *HCScript) Stop() {
	hc.Base.Stop()
}
