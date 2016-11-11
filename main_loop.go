package main

// main_loop is schedules healthchecks to be run as long as app_state.running is true
func main_loop(app_state *AppState, config *ConfigHead) {
	logger.Debug.Println("Entering main loop")

	app_state.running = true

	for app_state.running {
		for _, lb_pool := range config.LB_Pool {
			lb_pool.schedule_healthchecks()
		}
	}
}
