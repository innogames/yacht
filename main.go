package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/innogames/yacht/lbpool"
	"github.com/innogames/yacht/logger"
)

type AppState struct {
	//  commandline paramters
	verbose   bool
	no_action bool

	// configuration
	config_file string
	config      *map[string]interface{}

	// program operation
	stop_healthchecks chan bool
	program_running   bool
	wg                *sync.WaitGroup

	// LB Pools
	lbPools []*lbpool.LBPool
}

func (this *AppState) init_flags() {
	flag.StringVar(&this.config_file, "c", "/etc/iglb/iglb.json", "Location of confguration file")
	flag.BoolVar(&this.verbose, "v", false, "Be verbose, e.g. show every healhcheck")
	flag.BoolVar(&this.no_action, "n", false, "Do not perform any pfctl actions")
	flag.Parse()
}

func (this *AppState) init_signals() {
	c := make(chan os.Signal, 1)
	this.stop_healthchecks = make(chan bool)

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for {
			sig := <-c
			switch sig {
			case syscall.SIGINT:
				this.program_running = false
				this.stop_healthchecks <- true
			case syscall.SIGTERM:
				this.program_running = false
				this.stop_healthchecks <- true
			case syscall.SIGHUP:
				this.stop_healthchecks <- true
			}
		}
	}()
}

// loadConfig opens a JSON file, unmarshalls it and stores configuration in AppState.
func (this *AppState) loadConfig() {
	logger.Info.Printf("Loading configuration from %s", this.config_file)

	file, e := ioutil.ReadFile(this.config_file)
	if e != nil {
		logger.Error.Printf("Unable to open config file: %v\n", e)
		this.config = nil
	}

	this.config = new(map[string]interface{})
	json.Unmarshal(file, this.config)
}

// runLBPools materializes LB Pools from configuration stored in AppState.
// LBPools upon creation return only a channel that allows to stop them.
func (this *AppState) runLBPools() {

	// Ensure that configuration was loaded correctly
	if this.config == nil {
		logger.Error.Printf("No LB Pools found in config!")
		time.Sleep(1 * time.Second)
		return
	}

	logger.Debug.Printf("Starting LB Pools")
	if lb_pools, ok := (*this.config)["lbpools"].(map[string]interface{}); ok {
		for pool_k, pool_v := range lb_pools {
			pool_map := pool_v.(map[string]interface{})
			// For each LB Pool found in configuration file try to spawn a new
			// LBPool object both for IPv4 and IPv6. Nothing will be spanw if
			// LB Pool has no configured IP address for given protocol.
			if lbPool := lbpool.NewLBPool(this.wg, "ip4", pool_k, pool_map); lbPool != nil {
				this.lbPools = append(this.lbPools, lbPool)
			}
			if lbPool := lbpool.NewLBPool(this.wg, "ip6", pool_k, pool_map); lbPool != nil {
				this.lbPools = append(this.lbPools, lbPool)
			}
		}
	}
	logger.Debug.Printf("All LB Pools started")
}

// main_loop schedules healthchecks to be run as long as this.running is true
func (this *AppState) main_loop() {

	this.program_running = true

	logger.Debug.Println("Entering main loop")

	for this.program_running == true {

		// Prepare the wait group
		this.wg = new(sync.WaitGroup)

		// Load configuration and run loaded LB Pools.
		this.loadConfig()
		this.runLBPools()

		// Wait for a channel message which will terminate all running checks.
		select {
		case <-this.stop_healthchecks:
			for _, lbPool := range this.lbPools {
				lbPool.Stop()
			}
		}
		// Wait for healthchecks to be really finished.
		// This means: wait for wg counter to reach 0.
		this.wg.Wait()
	}
}

func main() {
	var appState AppState

	appState.init_flags()
	logger.InitLoggers(appState.verbose)

	logger.Info.Println("Yet Another Checking Health Tool starting")

	appState.init_signals()
	appState.main_loop()

	logger.Info.Println("Finished, good bye!")
}
