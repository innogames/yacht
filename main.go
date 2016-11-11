package main


func main() {
	var app_state AppState

	app_state.init_flags()
	app_state.init_logger()


	logger.Info.Println("Yet Another Checking Health Tool starting")

	app_state.init_signals()

	config := app_state.load_config()
	if config != nil {
		app_state.main_loop(config)
	} else {
		logger.Error.Println("Configuration file was not loaded or parsed")
	}

	logger.Info.Println("Finished, good bye!")
}
