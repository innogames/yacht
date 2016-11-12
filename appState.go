package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"

	"sync"
)

var logger struct {
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

type AppState struct {
	config_file       string
	verbose           bool
	no_action         bool
	stop_healthchecks chan bool
	program_running   bool
	wg                *sync.WaitGroup
}

func (app_state *AppState) init_flags() {
	flag.StringVar(&app_state.config_file, "c", "/etc/iglb/yahc.conf", "Location of confguration file")
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

func (app_state *AppState) init_logger() {
	var DebugWriter io.Writer

	if app_state.verbose {
		DebugWriter = os.Stderr
	} else {
		DebugWriter = ioutil.Discard
	}
	logger.Debug = log.New(DebugWriter, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	logger.Warning = log.New(os.Stderr, "WARNING: ", log.Ldate|log.Ltime)
	logger.Error = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// load_config loads configuration from given toml configuration file.
// After loading the file configuration is also compiled.
// It returns pointer to loaded and compiled configuration.
func (app_state *AppState) load_config() *ConfigHead {
	var conf ConfigHead

	logger.Info.Printf("Loading configuration from %s", app_state.config_file)

	md, err := toml.DecodeFile(app_state.config_file, &conf)
	if err != nil {
		logger.Error.Println(err)
		return nil
	}

	logger.Debug.Println("Loaded configuration:")
	conf.compile_config(md)

	return &conf
}

// main_loop is schedules healthchecks to be run as long as app_state.running is true
func (app_state *AppState) main_loop() {

	app_state.program_running = true

	logger.Debug.Println("Entering main loop")

	for app_state.program_running == true {
		config := app_state.load_config()

		if config == nil {
			logger.Error.Println("Configuration file was not loaded or it was impossible to parse")
			app_state.program_running = false
		} else {

			// Start all healthchecks.
			// They are goroutines and don't block this thread, hence the WaitGroup and Wait() below.
			app_state.wg = new(sync.WaitGroup)
			for i, _ := range config.LB_Pool {
				config.LB_Pool[i].run_healthchecks(app_state)
			}

			// Wait for a channel message which will terminate all running checks.
			select {
			case <-app_state.stop_healthchecks:
				// Send stop_healthchecks message to all healthchecks
				for i, _ := range config.LB_Pool {
					config.LB_Pool[i].stop_healthchecks()
				}
			}

			// Wait for healthchecks to be really finished.
			// They must finish talking to servers of loadbalancer before this program is terminated.
			// This is waiting for wg counter to reach 0.
			app_state.wg.Wait()
		}

		// Will garbage collector remove it?
		config = nil
	}
}
