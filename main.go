package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type AppState struct {
	config_file     string
	verbose         bool
	no_action       bool
	checks_running  bool
	program_running bool
	wg              *sync.WaitGroup
}

func init_flags(app_state *AppState) {
	flag.StringVar(&app_state.config_file, "c", "/etc/iglb/yahc.conf", "Location of confguration file")
	flag.BoolVar(&app_state.verbose, "v", false, "Be verbose, e.g. show every healhcheck")
	flag.BoolVar(&app_state.no_action, "n", false, "Do not perform any pfctl actions")
	flag.Parse()
}

func init_signals(app_state *AppState) {
	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for {
			sig := <-c
			switch sig {
			case syscall.SIGINT:
				app_state.checks_running = false
				app_state.program_running = false
			case syscall.SIGTERM:
				app_state.program_running = false
				app_state.checks_running = false
			case syscall.SIGHUP:
				app_state.checks_running = false
			}
		}
	}()

}

// main_loop is schedules healthchecks to be run as long as app_state.running is true
func main_loop(app_state *AppState) {

	app_state.program_running = true

	logger.Debug.Println("Entering main loop")

	for app_state.program_running == true {
		config := load_config(app_state)
		app_state.checks_running = true

		if config == nil {
			logger.Error.Println("Configuration file was not loaded or parsed")
		} else {
			app_state.wg = new(sync.WaitGroup)
			for i, _ := range config.LB_Pool {
				config.LB_Pool[i].run_healthchecks(app_state)
			}
			app_state.wg.Wait()
		}
		// Will garbage collector remove it?
		config = nil
	}
}

func main() {
	var app_state *AppState = new(AppState)

	init_flags(app_state)
	init_logger(app_state)

	logger.Info.Println("Yet Another Checking Health Tool starting")

	init_signals(app_state)
	main_loop(app_state)

	logger.Info.Println("Finished, good bye!")
}
