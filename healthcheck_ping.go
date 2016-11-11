package main

type HealthCheck_ping struct {
	HealthCheckBase
}

func (hc HealthCheck_ping) compile_config() {
	hc.HealthCheckBase.compile_config()
	logger.Info.Printf("       No specific parameters")
}
