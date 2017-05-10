package lbpool

import (
	"fmt"
	"github.com/innogames/yacht/logger"
	"sync"
)

// LBPool represents the object which receives the traffic and balances it between nodes.
type LBPool struct {
	// Properties
	name      string
	ipAddress string
	lbNodes   []*LBNode

	// Operation
	nodesStates NodesStates

	// Communication
	logPrefix string
	nodeChan  chan NodeStateMsg
	stopChan  chan bool
}

// NewLBPool is a object factory which creates new LBPool using configuration from JSON.
// The object's main goroutine is also started here.
func NewLBPool(proto string, name string, json map[string]interface{}) *LBPool {
	ipAddress := json[proto]
	if ipAddress == nil {
		return nil
	}

	// Initialize new LB Pool
	lbPool := new(LBPool)
	lbPool.stopChan = make(chan bool)
	lbPool.nodeChan = make(chan NodeStateMsg)
	lbPool.name = name + "_" + proto
	lbPool.ipAddress = ipAddress.(string)
	lbPool.logPrefix = fmt.Sprintf("lb_pool: %s ", lbPool.name)
	lbPool.nodesStates = NodesStates{}
	logger.Info.Printf(lbPool.logPrefix + "created")

	// Configuration of Healthchecks for this LB Pool will be passed to all nodes.
	// They will make their own HealthChecks from it.
	hcConfigs := json["healthchecks"]

	// Create LB Nodes for this LB Pool
	nodes := json["nodes"].(map[string]interface{})
	for nodeName, nodeConfig := range nodes {
		lbNode := newLBNode(lbPool.nodeChan, lbPool.logPrefix, proto, nodeName, nodeConfig.(map[string]interface{}), hcConfigs.([]interface{}))
		lbPool.lbNodes = append(lbPool.lbNodes, lbNode)
		lbPool.nodesStates[lbNode] = NodeState(NodeUnknown)
	}

	return lbPool
}

func (lbp *LBPool) poolLogic(nsm NodeStateMsg) {
	lbp.nodesStates.update(nsm)
	upNodes, allNodes := lbp.nodesStates.upNodes()
	logger.Info.Printf(lbp.logPrefix+"%d/%d nodes up", upNodes, allNodes)
}

// Run is the main loop of LB Pool.
func (lbp *LBPool) Run(wg *sync.WaitGroup) {

	// Start operation of all LB Nodes of this Pool.
	for _, lbNode := range lbp.lbNodes {
		go lbNode.run(wg)
	}

	for {
		select {
		case nsm := <-lbp.nodeChan:
			lbp.poolLogic(nsm)

		// Message from parent (main program): stop running.
		case <-lbp.stopChan:
			return
		}
	}
}

// Stop terminates operation of this LB Pool. It does it in proper order:
// first it terminates operation of all children and then of itself.
func (lbp *LBPool) Stop() {
	for _, lbNode := range lbp.lbNodes {
		lbNode.stop()
	}
	lbp.stopChan <- true
}
