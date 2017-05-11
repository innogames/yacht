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
	pfName    string

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
	logger.Info.Printf(lbPool.logPrefix + "created")

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
	lbp.wantedChanged = true
	//logger.Info.Printf(lbp.logPrefix+"%d/%d nodes up", upNodes, allNodes)
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
