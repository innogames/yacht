package main

import "github.com/BurntSushi/toml"

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
	schedule()
	compile_config()
}

func (hcc healthcheck_config) compile_config(md toml.MetaData) HealthCheck {
	var new_hc HealthCheck

	logger.Info.Printf("      HealthCheck '%s'", hcc.Type)
	switch hcc.Type {
	case "http":
		new_hc = new(HealthCheck_http)
	case "ping":
		new_hc = new(HealthCheck_ping)
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

func (hc HealthCheckBase) compile_config() {
	logger.Info.Printf("       General parameters:")
	logger.Info.Printf("       - Interval: '%d'", hc.Interval)
	logger.Info.Printf("       - MaxFfailed: '%d'", hc.Max_Failed)
}

func (hc HealthCheckBase) schedule() {
	logger.Debug.Printf("Scheduling HC %s", hc.Type)
}
