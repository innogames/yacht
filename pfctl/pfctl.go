package pfctl

import (
	"github.com/innogames/yacht/lbpool"
	"github.com/innogames/yacht/logger"
	"sync"
	"time"
)

// PFctl handles operations on pf
type PFctl struct {
	lbPools  []*lbpool.LBPool
	wg       *sync.WaitGroup
	active   bool
	stopChan chan bool
}

// NewPFctl creates new PFctl object
func NewPFctl(wg *sync.WaitGroup, lbPools []*lbpool.LBPool) *PFctl {
	pfctl := new(PFctl)
	pfctl.wg = wg
	pfctl.stopChan = make(chan bool)
	pfctl.lbPools = lbPools

	go pfctl.run()

	return pfctl
}

func (pfctl *PFctl) do() {
	for _, lbPool := range pfctl.lbPools {
		poolName, poolNodes, logPrefix := lbPool.GetWantedNodes()
		if poolNodes != nil {
			err := pfctlSyncTable(poolName, poolNodes)
			if err != nil {
				logger.Error.Printf(logPrefix + err.Error())
			}
		}
	}
}

func (pfctl *PFctl) run() {
	defer pfctl.wg.Done()
	pfctl.wg.Add(1)

	for {
		pfctl.do()
		select {
		case <-pfctl.stopChan:
			logger.Debug.Printf("received stopchan")
			return
		case <-time.After(time.Millisecond * time.Duration(100)):
		}
	}
}

// Stop terminates running pfctl operation
func (pfctl *PFctl) Stop() {
	logger.Debug.Printf("sending stopchan")
	pfctl.stopChan <- true
}
