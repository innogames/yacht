package lbpool

import (
	"fmt"
	"github.com/innogames/yacht/logger"
	"sync"
)

// LBPool represents the object which receives the traffic and balances it between nodes.
type LBPool struct {
	// Properties
	name           string
	ipAddress      string
	lbNodes        []*LBNode
	pfName         string
	minNodes       int
	maxNodes       int
	minNodesAction MinNodesAction

	// Operation
	sync.Mutex
	wantedNodes   []*LBNode
	wantedChanged bool

	// Communication
	logPrefix string
	nodeChan  chan NodeStateMsg
	stopChan  chan bool
}

// NewLBPool is a object factory which creates new LBPool using configuration from JSON.
// The object's main goroutine is also started here.
func NewLBPool(proto string, name string, json map[string]interface{}) *LBPool {
	// Skip Pools without IP Address
	ipAddress := json["ip"+proto]
	if ipAddress == nil {
		return nil
	}

	// Initialize new LB Pool
	lbPool := new(LBPool)
	lbPool.stopChan = make(chan bool)
	lbPool.name = name + "_" + proto
	lbPool.pfName = json["pf_name"].(string) + "_" + proto
	lbPool.ipAddress = ipAddress.(string)
	lbPool.logPrefix = fmt.Sprintf("lb_pool: %s ", lbPool.name)

	// Configure min and max nodes behaviour
	if minNodes, ok := json["min_nodes"].(float64); ok {
		lbPool.minNodes = int(minNodes)
	} else {
		lbPool.minNodes = 0
	}

	if maxNodes, ok := json["max_nodes"].(float64); ok {
		lbPool.maxNodes = int(maxNodes)
	} else {
		lbPool.maxNodes = 0
	}

	if lbPool.maxNodes > 0 && lbPool.maxNodes < lbPool.minNodes {
		lbPool.maxNodes = lbPool.minNodes
	}

	logger.Info.Printf(lbPool.logPrefix+"min %d max %d created", lbPool.minNodes, lbPool.maxNodes)

	// Configuration of Healthchecks for this LB Pool will be passed to all nodes.
	// They will make their own HealthChecks from it.
	hcConfigs := json["healthchecks"]

	// Create LB Nodes for this LB Pool
	nodes := json["nodes"].(map[string]interface{})
	for nodeName, nodeConfig := range nodes {
		lbNode := newLBNode(lbPool, lbPool.logPrefix, proto, nodeName, nodeConfig.(map[string]interface{}), hcConfigs.([]interface{}))
		lbPool.lbNodes = append(lbPool.lbNodes, lbNode)
	}

	return lbPool
}

// poolLogic handles adding and removing nodes.
// It is called from LB Node which should have already locked LB Pool struct.
func (lbp *LBPool) poolLogic(lbNode *LBNode) {
	// First check if state of all Nodes is known
	for _, lbn := range lbp.lbNodes {
		if lbn.state == NodeUnknown {
			return
		}
	}

	// Mark wanted set as dirty.
	lbp.wantedChanged = true

	var upNodes, forcedNodes, allNodes int
	var wantedNodes []*LBNode

	// Add nodes while satisfying maxNodes if it is set. Initial nodes don't
	// matter at this point becaue it is always better to replace them with
	// something that in fact has passed healthchecs.
	for _, lbn := range lbp.lbNodes {
		allNodes++
		if lbn.primary && lbn.state == NodeUp && (lbp.maxNodes == 0 || upNodes <= lbp.maxNodes) {
			wantedNodes = append(wantedNodes, lbn)
			upNodes++
		}
	}

	// Now satisfy minNodes depending on its configuration
	if lbp.minNodes > 0 && upNodes < lbp.minNodes {

		if lbp.minNodesAction == ForceDown {
			// ForceDown means that wantedNodes must be empty if not enough
			// up nodes are found.
			wantedNodes = []*LBNode{}
		} else if lbp.minNodesAction == ForceUp {
			// ForceUp means that any nodes must be added to wantedNodes
			// even if they are down. Start with node for which this function
			// was called. This is the the last one which was alive, so let's
			// not change loadbalancing.
			if lbNode.primary {
				wantedNodes = append(wantedNodes, lbNode)
				forcedNodes++
			}
			// Then try any other nodes.
			for _, lbn := range lbp.lbNodes {
				if lbn.primary && forcedNodes < lbp.minNodes {
					wantedNodes = append(wantedNodes, lbn)
					forcedNodes++
				}
			}
		} else if lbp.minNodesAction == BackupPool {
			for _, lbn := range lbp.lbNodes {
				if lbn.primary == false && lbn.state == NodeUp {
					wantedNodes = append(wantedNodes, lbn)
					forcedNodes++
				}
			}
		}
	}

	lbp.wantedNodes = wantedNodes
	logger.Info.Printf(lbp.logPrefix+"nodes: up %d forced %d min %d max %d all %d", upNodes, forcedNodes, lbp.minNodes, lbp.maxNodes, allNodes)
	for _, node := range wantedNodes {
		logger.Info.Printf(lbp.logPrefix+"lb_node: %s action: active", node.name)
	}
}

// GetWantedNodes returns information required to configure loadbalancing.
// If there was no change, it returns nil as list.
func (lbp *LBPool) GetWantedNodes() (string, []string) {
	lbp.Lock()
	if lbp.wantedChanged == false {
		return lbp.pfName, nil
	}
	lbp.wantedChanged = false
	var ret []string
	for _, lbn := range lbp.wantedNodes {
		ret = append(ret, lbn.ipAddress)
	}
	lbp.Unlock()
	return lbp.pfName, ret
}

// Run is the main loop of LB Pool.
func (lbp *LBPool) Run(wg *sync.WaitGroup) {

	// Start operation of all LB Nodes of this Pool.
	for _, lbNode := range lbp.lbNodes {
		go lbNode.run(wg)
	}

}

// Stop terminates operation of this LB Pool. It does it in proper order:
// first it terminates operation of all children and then of itself.
func (lbp *LBPool) Stop() {
	for _, lbNode := range lbp.lbNodes {
		lbNode.stop()
	}
}
