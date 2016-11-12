package main

import "github.com/BurntSushi/toml"

type LB_Pool struct {
	// Parameters filled in during initializiation
	Name            string
	IP_Address      string
	LB_Node         []LB_Node
	HealthCheckConf []healthcheck_config `toml:"healthcheck"`
}

func (lb_pool *LB_Pool) compile_config(md toml.MetaData) {
	logger.Info.Printf("  LB_Pool '%s'", lb_pool.Name)
	logger.Info.Printf("    Parameters:")
	logger.Info.Printf("    - IP Address: '%s'", lb_pool.IP_Address)

	for i, _ := range lb_pool.LB_Node {
		lb_pool.LB_Node[i].compile_config()
		logger.Info.Printf("     Healthchecks:")
		for _, hcc := range lb_pool.HealthCheckConf {
			// Append per-LB_Pool HealthChecks
			lb_pool.LB_Node[i].add_healthcheck(hcc.compile_config(md))
		}
		for _, hcc := range lb_pool.LB_Node[i].HealthCheckConf {
			// Append per-LB_Node HealthChecks
			lb_pool.LB_Node[i].add_healthcheck(hcc.compile_config(md))
		}
	}
}

func (lb_pool *LB_Pool) run_healthchecks(app_state *AppState) {
	for i, _ := range lb_pool.LB_Node {
		lb_pool.LB_Node[i].run_healthchecks(app_state)
	}
}

func (lb_pool *LB_Pool) stop_healthchecks() {
	for i, _ := range lb_pool.LB_Node {
		lb_pool.LB_Node[i].stop_healthchecks()
	}
}
