package main

type HealthCheck_script struct {
	HealthCheckBase
	Script  string
}

func (hc HealthCheck_script) compile_config() {
	hc.HealthCheckBase.compile_config()
	logger.Info.Printf("       Specific parameters:")
	logger.Info.Printf("       - Script: '%s'", hc.Script)
}