package main

type HealthCheck_script struct {
	HealthCheckBase
	Script string
}

func (hc HealthCheck_script) compile_config() {
	hc.HealthCheckBase.compile_config()
	logger.Info.Printf("       Specific parameters:")
	logger.Info.Printf("       - Script: '%s'", hc.Script)
}

func (hc *HealthCheck_script) run(app_state *AppState) {
	hc.HealthCheckBase.run(app_state)
}

func (hc *HealthCheck_script) stop() {
	hc.HealthCheckBase.stop()
}
