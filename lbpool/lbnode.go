package lbpool

import (
	"fmt"
	"github.com/innogames/yacht/healthcheck"
	"github.com/innogames/yacht/logger"
	"sync"
)

// LBNode represents one of nodes serving traffic in a loadbalancer. Here it stores
// values which are used by all Healthchecks.
type LBNode struct {
	// Properties
	name      string
	ipAddress string

	// Communication
	logPrefix    string
	stopChan     chan bool
	healthChecks []*healthcheck.HealthCheck
}

func newLBNode(wg *sync.WaitGroup, logPrefix string, proto string, name string, nodeConfig map[string]interface{}, hcConfigs []interface{}) *LBNode {
	if nodeConfig[proto] == nil {
		return nil
	}
	ipAddress := nodeConfig[proto].(string)

	// Initialize new LB Node
	lbNode := new(LBNode)
	lbNode.logPrefix = fmt.Sprintf(logPrefix+"lb_node: %s ", name)
	lbNode.stopChan = make(chan bool)

	logger.Info.Printf(lbNode.logPrefix + "created")

	// Run this node before healthchecks are created. They might send messages immediately!
	go lbNode.run()

	for _, hcConfig := range hcConfigs {
		hc := healthcheck.NewHealthCheck(wg, lbNode.logPrefix, hcConfig.(map[string]interface{}), ipAddress)
		lbNode.healthChecks = append(lbNode.healthChecks, hc)
	}

	return lbNode
}

// Run is the main loop of LB Node. It receives messages from parent and children.
func (lbn *LBNode) run() {
	for {
		select {
		// Message from parent (LB Pool): stop running.
		case <-lbn.stopChan:
			return
		}
	}
}

// Stop terminates operation of this LB Node. It does it in proper order:
// first it terminates operation of all children and then of itself.
func (lbn *LBNode) stop() {
	for _, hc := range lbn.healthChecks {
		(*hc).Stop()
	}
	lbn.stopChan <- true
}
