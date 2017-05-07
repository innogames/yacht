package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/innogames/yacht/lbpool"
	"github.com/innogames/yacht/logger"
)

type AppState struct {
	config_file       string
	verbose           bool
	no_action         bool
	stop_healthchecks chan bool
	program_running   bool
	wg                *sync.WaitGroup
}

func (app_state *AppState) init_flags() {
	flag.StringVar(&app_state.config_file, "c", "/etc/iglb/iglb.json", "Location of confguration file")
	flag.BoolVar(&app_state.verbose, "v", false, "Be verbose, e.g. show every healhcheck")
	flag.BoolVar(&app_state.no_action, "n", false, "Do not perform any pfctl actions")
	flag.Parse()
}

func (app_state *AppState) init_signals() {
	c := make(chan os.Signal, 1)
	app_state.stop_healthchecks = make(chan bool)

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for {
			sig := <-c
			switch sig {
			case syscall.SIGINT:
				app_state.stop_healthchecks <- true
				app_state.program_running = false
			case syscall.SIGTERM:
				app_state.stop_healthchecks <- true
				app_state.program_running = false
			case syscall.SIGHUP:
				app_state.stop_healthchecks <- true
			}
		}
	}()
}

// load_config loads configuration from given toml configuration file.
// After loading the file configuration is also compiled.
// It returns pointer to loaded and compiled configuration.
func (app_state *AppState) load_config() []*lbpool.LBPool {
	var all_pools []*lbpool.LBPool

	logger.Info.Printf("Loading configuration from %s", app_state.config_file)

	file, e := ioutil.ReadFile(app_state.config_file)
	if e != nil {
		logger.Error.Printf("Unable to open config file: %v\n", e)
		return nil
	}

	var json_config map[string]interface{}
	json.Unmarshal(file, &json_config)

	logger.Debug.Printf("Json loaded, parsing it")
	if lb_pools, ok := json_config["lbpools"].(map[string]interface{}); ok {
		for pool_k, pool_v := range lb_pools {
			pool_map := pool_v.(map[string]interface{})
			if pool_ip4 := lbpool.NewLBPool("ip4", pool_k, pool_map); pool_ip4 != nil {
				all_pools = append(all_pools, pool_ip4)
			}
			if pool_ip6 := lbpool.NewLBPool("ip6", pool_k, pool_map); pool_ip6 != nil {
				all_pools = append(all_pools, pool_ip6)
			}
		}
	}

	return all_pools
}

// main_loop schedules healthchecks to be run as long as app_state.running is true
func (app_state *AppState) main_loop() {

	app_state.program_running = true

	logger.Debug.Println("Entering main loop")

	for app_state.program_running == true {
		lb_pools := app_state.load_config()
		{
			app_state.wg = new(sync.WaitGroup)

			// Wait for a channel message which will terminate all running checks.
			select {
			case <-app_state.stop_healthchecks:
				// Send stop_healthchecks message to all healthchecks
				for i, _ := range lb_pools {
					lb_pools[i].StopHealthChecks()
				}
			}

			// Wait for healthchecks to be really finished.
			// They must finish talking to servers of loadbalancer before this program is terminated.
			// This is waiting for wg counter to reach 0.
			app_state.wg.Wait()
		}

		// Will garbage collector remove it?
		lb_pools = nil
	}
}

func main() {
	var app_state AppState

	app_state.init_flags()
	logger.InitLoggers(app_state.verbose)

	logger.Info.Println("Yet Another Checking Health Tool starting")

	app_state.init_signals()
	app_state.main_loop()

	logger.Info.Println("Finished, good bye!")
}
