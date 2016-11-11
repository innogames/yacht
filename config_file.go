package main

import "github.com/BurntSushi/toml"

type ConfigHead struct {
	LB_Pool []LB_Pool
}

// compile_config fills in some variables which were not loaded from configuration file.
// It goes through list of loaded LB Nodes and asks each one to compile its own config.
func (config ConfigHead) compile_config(md toml.MetaData) {
	logger.Info.Printf("Compiling configuration of yacht")
	for _, lb_pool := range config.LB_Pool {
		lb_pool.compile_config(md)
	}
}

// load_config loads configuration from given toml configuration file.
// After loading the file configuration is also compiled.
// It returns pointer to loaded and compiled configuration.
func load_config(app_state *AppState) *ConfigHead {
	var conf ConfigHead

	logger.Info.Printf("Loading configuration from %s", app_state.config_file)

	md, err := toml.DecodeFile(app_state.config_file, &conf)
	if err != nil {
		logger.Error.Println(err)
		return nil
	}

	logger.Debug.Println("Loaded configuration:")
	conf.compile_config(md)

	return &conf
}
