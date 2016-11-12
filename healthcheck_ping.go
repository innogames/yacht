package main

type HealthCheck_ping struct {
	HealthCheckBase
}

func (hc *HealthCheck_ping) compile_config() {
	hc.HealthCheckBase.compile_config()
	logger.Info.Printf("       No specific parameters")
}

func (hc *HealthCheck_ping) run(app_state *AppState) {
	hc.HealthCheckBase.run(app_state)
}

func (hc *HealthCheck_ping) stop() {
	hc.HealthCheckBase.stop()
}
