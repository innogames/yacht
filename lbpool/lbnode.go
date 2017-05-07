package lbpool

import (
	"github.com/innogames/yacht/healthcheck"
	"github.com/innogames/yacht/logger"
	"sync"
)

type LBNode struct {
	// Properties
	name      string
	ipAddress string

	// Communication
	stopChan     chan bool
	healthChecks []*healthcheck.HealthCheck
}

func newLBNode(wg *sync.WaitGroup, proto string, name string, nodeConfig map[string]interface{}, hcConfigs []interface{}) *LBNode {
	ipAddress := nodeConfig[proto]
	if ipAddress == nil {
		return nil
	}

	// Initialize new LB Node
	lbNode := new(LBNode)
	lbNode.stopChan = make(chan bool)
	logger.Info.Printf("lb_node: %s, ip_ddress: %s, action: create", name, ipAddress)

	// Run this node before healthchecks are created. They might send messages immediately!
	go lbNode.run()

	for _, hcConfig := range hcConfigs {
		hc := healthcheck.NewHealthCheck(wg, hcConfig.(map[string]interface{}))
		lbNode.healthChecks = append(lbNode.healthChecks, hc)
	}

	return lbNode
}

// Run is the main loop of LB Node. It receives messages from parent and children.
func (this *LBNode) run() {
	for {
		select {
		// Message from parent (LB Pool): stop running.
		case <-this.stopChan:
			return
		}
	}
}

// Stop terminates operation of this LB Node. It does it in proper order:
// first it terminates operation of all children and then of itself.
func (this *LBNode) stop() {
	for _, hc := range this.healthChecks {
		(*hc).Stop()
	}
	this.stopChan <- true
}
