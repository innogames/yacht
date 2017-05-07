package lbpool

import (
	"github.com/innogames/yacht/logger"
	"sync"
)

type LBPool struct {
	// Properties
	name      string
	ipAddress string
	lbNodes   []*LBNode

	// Communication
	wg       *sync.WaitGroup
	stopChan chan bool
}

func NewLBPool(wg *sync.WaitGroup, proto string, name string, json map[string]interface{}) *LBPool {
	ipAddress := json[proto]
	if ipAddress == nil {
		return nil
	}

	// Initialize new LB Pool
	lbPool := new(LBPool)
	lbPool.stopChan = make(chan bool)
	lbPool.name = name
	lbPool.ipAddress = ipAddress.(string)
	logger.Info.Printf("lb_pool: %s, ip_address: %s, action: create", lbPool.name, lbPool.ipAddress)

	// Configuration of Healthchecks for this LB Pool will be passed to all nodes.
	// They will make their own HealthChecks from it.
	hcConfigs := json["healthchecks"]

	// Run this pool before nodes are created. They might send messages immediately!
	go lbPool.run()

	// Create LB Nodes for this LB Pool
	nodes := json["nodes"].(map[string]interface{})
	for nodeName, nodeConfig := range nodes {
		lbnode := newLBNode(wg, proto, nodeName, nodeConfig.(map[string]interface{}), hcConfigs.([]interface{}))
		lbPool.lbNodes = append(lbPool.lbNodes, lbnode)
	}

	return lbPool
}

// Run is the main loop of LB Pool.
func (this *LBPool) run() {

	for {
		select {
		// Message from parent (main program): stop running.
		case <-this.stopChan:
			return
		}
	}
}

// Stop terminates operation of this LB Pool. It does it in proper order:
// first it terminates operation of all children and then of itself.
func (this *LBPool) Stop() {
	for _, lbNode := range this.lbNodes {
		lbNode.stop()
	}
	this.stopChan <- true
}
