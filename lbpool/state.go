package lbpool

// NodeStateMsg represents if node is usable for serving traffic or not.
type NodeStateMsg struct {
	state  NodeState
	lbNode *LBNode
}

// NodeState represents all states a LB Node can be in, see below for enum values.
type NodeState int

const (
	// NodeUnknown is the default value
	NodeUnknown NodeState = iota
	// NodeDown can't serve traffic as it has some of its HCs failed.
	NodeDown
	// NodeUp has all HCs passed and thus can serve traffic
	NodeUp
)

// NodesStates is used to store state of many LN Nodes.
type NodesStates map[*LBNode]NodeState

func (ns *NodesStates) upNodes() (int, int) {
	var allNodes int
	var upNodes int
	for _, state := range *ns {
		if state == NodeUp {
			upNodes++
		}
		allNodes++
	}
	return upNodes, allNodes
}

func (ns *NodesStates) update(nsc NodeStateMsg) {
	(*ns)[nsc.lbNode] = nsc.state
}
