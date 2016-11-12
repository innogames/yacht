package main

type LB_Node struct {
	// Variables filled in during config file loading
	Name            string
	IP_Address      string
	HealthCheckConf []healthcheck_config `toml:"healthcheck"`
	HealthChecks    []HealthCheck
}

//func (lb_node *LB_Node) compile_config(health_checks_conf []HealthCheck) {
func (lb_node *LB_Node) compile_config() {
	logger.Info.Printf("    LB_Node '%s'", lb_node.Name)
	logger.Info.Printf("     Parameters")
	logger.Info.Printf("     - IP Address: '%s'", lb_node.IP_Address)
}

func (lb_node *LB_Node) add_healthcheck(hc HealthCheck) {
	lb_node.HealthChecks = append(lb_node.HealthChecks, hc)
}

func (lb_node *LB_Node) run_healthchecks(app_state *AppState) {
	for i, _ := range lb_node.HealthChecks {
		lb_node.HealthChecks[i].run(app_state)
	}
}

func (lb_node *LB_Node) stop_healthchecks() {
	for i, _ := range lb_node.HealthChecks {
		lb_node.HealthChecks[i].stop()
	}
}
