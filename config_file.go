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


