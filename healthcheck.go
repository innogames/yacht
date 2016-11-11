package main

import (
	"time"

	"github.com/BurntSushi/toml"
)

type healthcheck_config struct {
	Type   string
	Config toml.Primitive
}

type HealthCheckBase struct {
	// Variables filled in during config file loading
	Type             string
	Interval         int
	Max_Failed       int
	Min_Nodes        int
	Min_Nodes_Action string
	Max_Nodes        int

	// Variables filled in later
	running bool
}

type HealthCheck interface {
	run(app_state *AppState)
	compile_config()
}

func (hcc *healthcheck_config) compile_config(md toml.MetaData) HealthCheck {
	var new_hc HealthCheck

	logger.Info.Printf("      HealthCheck '%s'", hcc.Type)
	switch hcc.Type {
	case "http":
		new_hc = new(HealthCheck_http)
	case "ping":
		new_hc = new(HealthCheck_ping)
	case "script":
		new_hc = new(HealthCheck_script)
	default:
		logger.Error.Printf("          Unknown HealthCheck type %q", hcc.Type)
		return nil
	}
	if err := md.PrimitiveDecode(hcc.Config, new_hc); err != nil {
		logger.Error.Printf("          Unable to parse HealthCheck's config:")
		logger.Error.Println(err)
		return nil
	}
	new_hc.compile_config()
	return new_hc
}

func (hc *HealthCheckBase) compile_config() {
	logger.Info.Printf("       General parameters:")
	logger.Info.Printf("       - Interval: '%d'", hc.Interval)
	logger.Info.Printf("       - MaxFfailed: '%d'", hc.Max_Failed)
}

func (hc *HealthCheckBase) run(app_state *AppState) {
	app_state.wg.Add(1)
	go func() {
		for app_state.checks_running == true {
			logger.Info.Printf("HC %v running", hc)
			time.Sleep(1 * time.Second)
			logger.Info.Printf("HC %v finished", hc)
			time.Sleep(1 * time.Second)
		}
		app_state.wg.Done()
	}()
}
