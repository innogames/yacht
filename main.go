package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
)

type AppState struct {
	config_file string
	verbose     bool
	no_action   bool
	running     bool
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
				app_state.running = false
			case syscall.SIGTERM:
				app_state.running = false
			case syscall.SIGHUP:
				logger.Warning.Println("Reloading not supported (yet)")
			}
		}
	}()

}

func main() {
	var app_state *AppState = new(AppState)

	init_flags(app_state)
	init_logger(app_state)

	logger.Info.Println("Yet Another Checking Health Tool starting")

	init_signals(app_state)

	config := load_config(app_state)
	if config != nil {
		main_loop(app_state, config)
	} else {
		logger.Error.Println("Configuration file was not loaded or parsed")
	}

	logger.Info.Println("Finished, good bye!")
}
