package main

import "github.com/BurntSushi/toml"

type ConfigHead struct {
	LB_Pool []LB_Pool
}

// compile_config fills in some variables which were not loaded from configuration file.
// It goes through list of loaded LB Nodes and asks each one to compile its own config.
func (config ConfigHead) compile_config(md toml.MetaData) {
	logger.Info.Printf("Compiling configuration of yacht")
	for i, _ := range config.LB_Pool {
		config.LB_Pool[i].compile_config(md)
	}
}


