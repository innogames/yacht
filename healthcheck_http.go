package main

type HealthCheck_http struct {
	HealthCheckBase
	Host     string
	Url      string
	OK_codes []int
}

func (hc *HealthCheck_http) compile_config() {
	hc.HealthCheckBase.compile_config()
	logger.Info.Printf("       Specific parameters:")
	logger.Info.Printf("       - Host: '%s'", hc.Host)
	logger.Info.Printf("       - URL: '%s'", hc.Url)
	logger.Info.Printf("       - OK Codes: %v", hc.OK_codes)
}

func (hc *HealthCheck_http) run(app_state *AppState) {
	hc.HealthCheckBase.run(app_state)
}

func (hc *HealthCheck_http) stop() {
	hc.HealthCheckBase.stop()
}
