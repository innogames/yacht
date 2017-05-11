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

// AppState holds some variables which otherwise would be considered global.
type AppState struct {
	//  commandline paramters
	verbose  bool
	noAction bool

	// configuration
	configFile string
	config     *map[string]interface{}

	// program operation
	stopHealthChecks chan bool
	programRunning   bool
	wg               *sync.WaitGroup

	// LB Pools
	lbPools []*lbpool.LBPool
}

func (appState *AppState) initFlags() {
	flag.StringVar(&appState.configFile, "c", "/etc/iglb/iglb.json", "Location of confguration file")
	flag.BoolVar(&appState.verbose, "v", false, "Be verbose, e.g. show every healhcheck")
	flag.BoolVar(&appState.noAction, "n", false, "Do not perform any pfctl actions")
	flag.Parse()
}

func (appState *AppState) initSignals() {
	c := make(chan os.Signal, 1)
	appState.stopHealthChecks = make(chan bool)

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for {
			sig := <-c
			switch sig {
			case syscall.SIGINT:
				appState.programRunning = false
				appState.stopHealthChecks <- true
			case syscall.SIGTERM:
				appState.programRunning = false
				appState.stopHealthChecks <- true
			case syscall.SIGHUP:
				appState.stopHealthChecks <- true
			}
		}
	}()
}

// loadConfig opens a JSON file, unmarshalls it and stores configuration in AppState.
func (appState *AppState) loadConfig() {
	logger.Info.Printf("Loading configuration from %s", appState.configFile)

	file, e := ioutil.ReadFile(appState.configFile)
	if e != nil {
		logger.Error.Printf("Unable to open config file: %v\n", e)
		appState.config = nil
	}

	appState.config = new(map[string]interface{})
	json.Unmarshal(file, appState.config)
}

// runLBPools materializes LB Pools from configuration in AppState.
// Each of LB Pools will then run as a goroutine.
func (appState *AppState) runLBPools() {

	// Ensure that configuration was loaded correctly
	if appState.config == nil {
		logger.Error.Printf("No LB Pools found in config!")
		time.Sleep(1 * time.Second)
		return
	}

	logger.Debug.Printf("Creating and starting LB Pools")
	if lbPools, ok := (*appState.config)["lbpools"].(map[string]interface{}); ok {
		for poolName, poolConfig := range lbPools {
			poolConfigMap := poolConfig.(map[string]interface{})
			// For each LB Pool found in configuration file try to spawn a new
			// LBPool object both for IPv4 and IPv6. Nothing will be spanw if
			// LB Pool has no configured IP address for given protocol.
			if lbPool := lbpool.NewLBPool("4", poolName, poolConfigMap); lbPool != nil {
				appState.lbPools = append(appState.lbPools, lbPool)
				go lbPool.Run(appState.wg)
			}
			if lbPool := lbpool.NewLBPool("6", poolName, poolConfigMap); lbPool != nil {
				appState.lbPools = append(appState.lbPools, lbPool)
				go lbPool.Run(appState.wg)
			}
		}
	}
	logger.Debug.Printf("All LB Pools started")
}

// mainLoop of the whole program. It loads configuration, creates all LB Pools,
// runs them and awaits them to finish working. After that loads the config again
// and repeats the whole process.
func (appState *AppState) mainLoop() {

	appState.programRunning = true

	logger.Debug.Println("Entering main loop")

	for appState.programRunning == true {

		// Prepare the wait group
		appState.wg = new(sync.WaitGroup)

		// Load configuration and run loaded LB Pools.
		appState.loadConfig()
		appState.runLBPools()

		// Wait for a channel message which will terminate all running checks.
		select {
		case <-appState.stopHealthChecks:
			for _, lbPool := range appState.lbPools {
				lbPool.Stop()
			}
		}
		// Wait for healthchecks to be really finished.
		// This means: wait for wg counter to reach 0.
		appState.wg.Wait()
	}
}

func main() {
	var appState AppState

	appState.initFlags()
	logger.InitLoggers(appState.verbose)

	logger.Info.Println("Yet Another Checking Health Tool starting")

	appState.initSignals()
	appState.mainLoop()

	logger.Info.Println("Finished, good bye!")
}
