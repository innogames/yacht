package lbpool

import (
	"github.com/innogames/yacht/healthcheck"
	"github.com/innogames/yacht/logger"
	"sync"
)

type LBNode struct {
	name         string
	ipAddress    string
	healthChecks []*healthcheck.HealthCheck
}

func newLBNode(proto string, name string, nodeConfig map[string]interface{}, hcConfigs []interface{}) *LBNode {
	ipAddress := nodeConfig[proto]
	if ipAddress == nil {
		return nil
	}

	logger.Info.Printf("lb_node: %s, ip_ddress: %s, action: create", name, ipAddress)

	lbnode := new(LBNode)

	for _, hcConfig := range hcConfigs {
		hc := healthcheck.NewHealthCheck(hcConfig.(map[string]interface{}))
		lbnode.healthChecks = append(lbnode.healthChecks, hc)
	}

	return lbnode
}

func (this LBNode) runHealthChecks(wg *sync.WaitGroup) {
	for _, v := range this.healthChecks {
		(*v).Run(wg)
	}
}

func (this LBNode) stopHealthChecks() {
	for _, v := range this.healthChecks {
		(*v).Stop()
	}
}
