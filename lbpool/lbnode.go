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
	// Configuration
	name      string
	ipAddress string

	// Operation
	okHCs int // Counts passed HCs. When it reaches 0, the node is down.

	// Communication
	logPrefix    string
	stopChan     chan bool
	hcChan       chan bool // Channel over which HCs will tell us their hard state
	healthChecks []*healthcheck.HealthCheck
}

func newLBNode(logPrefix string, proto string, name string, nodeConfig map[string]interface{}, hcConfigs []interface{}) *LBNode {
	if nodeConfig[proto] == nil {
		return nil
	}
	ipAddress := nodeConfig[proto].(string)

	// Initialize new LB Node
	lbNode := new(LBNode)
	lbNode.logPrefix = fmt.Sprintf(logPrefix+"lb_node: %s ", name)
	lbNode.stopChan = make(chan bool)
	lbNode.hcChan = make(chan bool)

	logger.Info.Printf(lbNode.logPrefix + "created")

	// First we create HCs. They are allowed to fail creation for example because
	// of unknow type or other trouble reading their configuration.
	for _, hcConfig := range hcConfigs {
		hc := healthcheck.NewHealthCheck(lbNode.hcChan, lbNode.logPrefix, hcConfig.(map[string]interface{}), ipAddress)
		if hc != nil {
			lbNode.healthChecks = append(lbNode.healthChecks, hc)
		}
	}

	// If no HCs were added because of bad configuraion or because not being
	// configured at all, add a simple dummy HC that always returns hcGood.
	// This makes Pools without HCs always have all Nodes forced up.
	dummyConfig := map[string]interface{}{
		"type":   "dummy",
		"result": healthcheck.HCGood,
	}
	if len(lbNode.healthChecks) == 0 {
		hc := healthcheck.NewHealthCheck(lbNode.hcChan, lbNode.logPrefix, dummyConfig, ipAddress)
		lbNode.healthChecks = append(lbNode.healthChecks, hc)
	}

	return lbNode
}

// Run is the main loop of LB Node. It receives messages from parent and children.
func (lbn *LBNode) run(wg *sync.WaitGroup) {

	for _, hc := range lbn.healthChecks {
		go (*hc).Run(wg)
	}

	for {
		select {
		// Message from one of Healthchecks about reaching a hard state.
		case hcState := <-lbn.hcChan:
			if hcState == true {
				lbn.okHCs++
			} else {
				lbn.okHCs--
			}
			if lbn.okHCs == len(lbn.healthChecks) {
				logger.Info.Printf(lbn.logPrefix + "up")
			} else {
				logger.Info.Printf(lbn.logPrefix + "down")
			}
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
