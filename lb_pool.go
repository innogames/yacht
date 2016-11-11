package main

import (
	"github.com/BurntSushi/toml"
//	"github.com/Syncbak-Git/nsca"
)

type LB_Pool struct {
	// Parameters filled in during initializiation
	Name            string
	IP_Address      string
	LB_Node         []LB_Node
	HealthCheckConf []healthcheck_config `toml:"healthcheck"`
}

func (lb_pool LB_Pool) compile_config(md toml.MetaData) {
	logger.Info.Printf("  LB_Pool '%s'", lb_pool.Name)
	logger.Info.Printf("    Parameters:")
	logger.Info.Printf("    - IP Address: '%s'", lb_pool.IP_Address)

	for _, lb_node := range lb_pool.LB_Node {
		lb_node.compile_config()
		logger.Info.Printf("     Healthchecks:")
		for _, hcc := range lb_pool.HealthCheckConf {
			// Append per-LB_Pool HealthChecks
			lb_node.HealthChecks = append(lb_node.HealthChecks, hcc.compile_config(md))
		}
		for _, hcc := range lb_node.HealthCheckConf {
			// Append per-LB_Node HealthChecks
			lb_node.HealthChecks = append(lb_node.HealthChecks, hcc.compile_config(md))
		}
	}
}

func (lb_pool LB_Pool) schedule_healthchecks() {
	//	for _, lb_node := range lb_pool.LB_Node {
	//		lb_node.schedule_healthchecks()
	///	}
}

func (lb_pool LB_Pool) notifyNagios() {

/*
	message := nsca.Message{State: nsca.STATE_OK, Host: "LB_pool_name", Service: "check_loadbalancer", Message: "All ok"}
	nagios_connection.messages <- message
 */


	//logger.Debug.Printf("Sending NSCA %s", hc.Type)

}
