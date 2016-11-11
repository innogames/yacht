package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"io"
	"io/ioutil"
	"log"
	"github.com/BurntSushi/toml"
//	"time"
//	"github.com/Syncbak-Git/nsca"
)

var logger struct {
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

/*var nagios_connection struct {
	messages   	chan nsca.Message
	quit 		chan interface{}
}*/

type AppState struct {
	config_file string
	verbose     bool
	no_action   bool
	running     bool
}

func (app_state * AppState) init_flags() {
	flag.StringVar(&app_state.config_file, "c", "/etc/iglb/yahc.conf", "Location of confguration file")
	flag.BoolVar(&app_state.verbose, "v", false, "Be verbose, e.g. show every healhcheck")
	flag.BoolVar(&app_state.no_action, "n", false, "Do not perform any pfctl actions")
	flag.Parse()
}

func (app_state * AppState) init_signals() {
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

func (app_state * AppState) init_logger() {
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
func (app_state * AppState) load_config() *ConfigHead {
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
func (app_state * AppState) main_loop(config *ConfigHead) {
	logger.Debug.Println("Entering main loop")

	app_state.running = true

	for app_state.running {
		for _, lb_pool := range config.LB_Pool {
			lb_pool.schedule_healthchecks()
		}
	}
}

/*func (app_state *AppState) init_nsca() {
	// This needs to come from config
	serverInfo := nsca.ServerInfo{Host: "nagios", Port: "4567", Timeout: time.Duration(1)*time.Second}
	nsca.RunEndpoint(serverInfo, nagios_connection.quit, *nagios_connection.messages)
}*/
