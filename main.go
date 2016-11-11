package main

func main() {
	var app_state AppState

	app_state.init_flags()
	app_state.init_logger()

	logger.Info.Println("Yet Another Checking Health Tool starting")

	app_state.init_signals()
	app_state.main_loop()

	logger.Info.Println("Finished, good bye!")
}
