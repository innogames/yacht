package main

type LB_Node struct {
	// Variables filled in during config file loading
	Name            string
	IP_Address      string
	HealthCheckConf []healthcheck_config `toml:"healthcheck"`
	HealthChecks    []HealthCheck
}

//func (lb_node *LB_Node) compile_config(health_checks_conf []HealthCheck) {
func (lb_node LB_Node) compile_config() {
	logger.Info.Printf("    LB_Node '%s'", lb_node.Name)
	logger.Info.Printf("     Parameters")
	logger.Info.Printf("     - IP Address: '%s'", lb_node.IP_Address)
}

func (lb_node LB_Node) schedule_healthchecks() {
	//	for _, hc := range lb_node.healthchecks {
	//		hc.schedule()
	//	}
}
