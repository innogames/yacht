package lbpool

// MinNodesAction tells what to do when there are not enough nodes in Pool
type MinNodesAction int

const (
	// ForceUp means that last nodes will not be removed or some even dead
	// nodes will be added to Pool after Yacht restart.
	ForceUp MinNodesAction = iota
	// ForceDown removes every other node from Pool
	ForceDown
	// BackupPool switches traffic to backup nodes in this Pool
	BackupPool
)
