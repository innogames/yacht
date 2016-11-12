package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"sync"
)

var logger struct {
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

type AppState struct {
	config_file     string
	verbose         bool
	no_action       bool
	checks_running  bool
	program_running bool
	wg              *sync.WaitGroup
}

func (app_state *AppState) init_flags() {
	flag.StringVar(&app_state.config_file, "c", "/etc/iglb/yahc.conf", "Location of confguration file")
	flag.BoolVar(&app_state.verbose, "v", false, "Be verbose, e.g. show every healhcheck")
	flag.BoolVar(&app_state.no_action, "n", false, "Do not perform any pfctl actions")
	flag.Parse()
}

func (app_state *AppState) init_signals() {
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
