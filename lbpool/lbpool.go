package lbpool

import (
	"github.com/innogames/yacht/logger"
	"sync"
)

type LBPool struct {
	name      string
	ipAddress string
	lbNodes   []*LBNode
}

func NewLBPool(proto string, name string, json JSONMap) *LBPool {
	ipAddress := json[proto]
	if ipAddress == nil {
		return nil
	}

	// Initialize a new LB Pool
	lbpool := new(LBPool)
	lbpool.name = name
	lbpool.ipAddress = ipAddress.(string)
	logger.Info.Printf("lb_pool: %s, ip_address: %s, action: create", lbpool.name, lbpool.ipAddress)

	// Configuration of Healthchecks for this LB Pool will be passed to all nodes.
	// They will make their own HealthChecks from it.
	hcConfigs := json["healthchecks"]

	// Create LB Nodes for this LB Pool
	nodes := json["nodes"].(map[string]interface{})
	for nodeName, nodeConfig := range nodes {
		lbnode := newLBNode(proto, nodeName, nodeConfig.(map[string]interface{}), hcConfigs.([]interface{}))
		lbpool.lbNodes = append(lbpool.lbNodes, lbnode)
	}

	return lbpool
}

func (this *LBPool) RunHealthChecks(wg *sync.WaitGroup) {
	for _, v := range this.lbNodes {
		v.runHealthChecks(wg)
	}
}

func (this *LBPool) StopHealthChecks() {
	for _, v := range this.lbNodes {
		v.stopHealthChecks()
	}
}
