package healthcheck

import (
	"github.com/innogames/yacht/logger"
	"sync"
)

// HCHttp stores all properties of a HTTP or HTTPS healthcheck.
type HCHttp struct {
	HealthCheckBase
	host    string
	url     string
	okCodes []int
}

// NewHCHttp creates new HTTP or HTTPs healthcheck struct and populates it with data from Json config
func newHCHttp(json JSONMap) *HCHttp {
	hc := new(HCHttp)
	hc.hcType = json["type"].(string)
	if host, ok := json["host"].(string); ok {
		hc.host = host
	}
	if url, ok := json["url"].(string); ok {
		hc.url = url
	}
	if codes, ok := json["ok_codes"].([]int); ok {
		hc.okCodes = codes
	}

	logger.Info.Printf("healthcheck: %s, url: %s", hc.hcType, hc.url)
	return hc
}

func (this HCHttp) Run(wg *sync.WaitGroup) {
	this.HealthCheckBase.Run(wg)
}

func (this HCHttp) Stop() {
	this.HealthCheckBase.Stop()
}
