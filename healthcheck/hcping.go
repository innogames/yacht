package healthcheck

import (
	"github.com/innogames/yacht/logger"
	"sync"
)

// HCPing stores all properties of a ping healthcheck. In fact it stores nothing
// because all required parameters are stored in HCbase.
type HCPing struct {
	HealthCheckBase
}

// NewHCPing creates new ping healthcheck struct and populates it with data from Json config.
func newHCPing(json JSONMap) *HCPing {
	hc := new(HCPing)
	hc.HealthCheckBase.configure(json)
	hc.hcType = json["type"].(string)
	logger.Info.Printf("healthcheck: %s", hc.hcType)
	return hc
}

func (this *HCPing) run(wg *sync.WaitGroup) {
	this.HealthCheckBase.run(wg)
}

func (this *HCPing) Stop() {
	this.HealthCheckBase.Stop()
}
