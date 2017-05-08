package healthcheck

import (
	"github.com/innogames/yacht/logger"
	"sync"
)

// HCHttp stores all properties of a HTTP or HTTPS healthcheck.
type HCHttp struct {
	Base
	host    string
	url     string
	okCodes []int
}

// NewHCHttp creates new HTTP or HTTPs healthcheck struct and populates it with data from Json config
func newHCHttp(json JSONMap) *HCHttp {
	hc := new(HCHttp)
	hc.Base.configure(json)
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

// run starts operation of this healthcheck, in fact it calls the Base class.
func (hc *HCHttp) run(wg *sync.WaitGroup) {
	hc.Base.run(wg)
}

// Stop terminates this healthcheck, in fact it calls the Base class.
func (hc *HCHttp) Stop() {
	hc.Base.Stop()
}
