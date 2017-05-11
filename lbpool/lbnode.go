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
	hcsResults healthcheck.HCsResults
	state      NodeState
	reason     NodeReason
	checked    bool // stores information if this node was ever checked

	// Communication
	logPrefix    string
	lbPool       *LBPool
	stopChan     chan bool
	hcChan       chan healthcheck.HCResultMsg // incoming channel over which HCs will tell us their hard state
	healthChecks []healthcheck.HealthCheck
}

func newLBNode(lbPool *LBPool, logPrefix string, proto string, name string, nodeConfig map[string]interface{}, hcConfigs []interface{}) *LBNode {
	// Skip Nodes without IP Address
	if nodeConfig["ip"+proto] == nil {
		return nil
	}

	// Initialize new LB Node
	lbNode := new(LBNode)
	lbNode.lbPool = lbPool
	lbNode.ipAddress = nodeConfig["ip"+proto].(string)
	lbNode.logPrefix = fmt.Sprintf(logPrefix+"lb_node: %s ", name)
	lbNode.stopChan = make(chan bool)
	lbNode.hcChan = make(chan healthcheck.HCResultMsg)
	lbNode.hcsResults = healthcheck.HCsResults{}

	logger.Info.Printf(lbNode.logPrefix + "created")

	// First we create HCs. They are allowed to fail creation for example because
	// of unknow type or other trouble reading their configuration.
	for _, hcConfig := range hcConfigs {
		hc := healthcheck.NewHealthCheck(lbNode.hcChan, lbNode.logPrefix, hcConfig.(map[string]interface{}), lbNode.ipAddress)
		if hc != nil {
			lbNode.healthChecks = append(lbNode.healthChecks, hc)
			lbNode.hcsResults[hc] = healthcheck.HCResult(healthcheck.HCUnknown)
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
		hc := healthcheck.NewHealthCheck(lbNode.hcChan, lbNode.logPrefix, dummyConfig, lbNode.ipAddress)
		lbNode.healthChecks = append(lbNode.healthChecks, hc)
		lbNode.hcsResults[hc] = healthcheck.HCResult(healthcheck.HCUnknown)
	}

	return lbNode
}

// nodeLogic is trigerred when state of any of HCs of this Node changes.
// Access to this LB Node must be protected because it can be accessed from
// another node's run() for lbPool.poolLogic()
func (lbn *LBNode) nodeLogic(hcrm healthcheck.HCResultMsg) {
	lbn.lbPool.Lock()

	lbn.checked = true
	lbn.hcsResults.Update(hcrm)
	goodHCs, allHCs, unknownHCs := lbn.hcsResults.GoodHCs()

	// Do not perform any actions untill all HCs report at least once!
	if unknownHCs == 0 {
		if goodHCs == allHCs && lbn.state != NodeUp {
			logger.Info.Printf(lbn.logPrefix+"%d/%d healthchecks good action: up", goodHCs, allHCs)
			lbn.state = NodeUp
			lbn.lbPool.poolLogic(lbn)
		} else if goodHCs != allHCs && lbn.state != NodeDown {
			logger.Info.Printf(lbn.logPrefix+"%d/%d healthchecks good action: down", goodHCs, allHCs)
			lbn.state = NodeDown
			lbn.lbPool.poolLogic(lbn)
		}
	}
	lbn.lbPool.Unlock()
}

// Run is the main loop of LB Node. It receives messages from parent and children.
func (lbn *LBNode) run(wg *sync.WaitGroup) {

	for _, hc := range lbn.healthChecks {
		go hc.Run(wg)
	}

	for {
		select {
		// Message from one of Healthchecks about reaching a hard state.
		case hcrm := <-lbn.hcChan:
			lbn.nodeLogic(hcrm)
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
		hc.Stop()
	}
	lbn.stopChan <- true
}
