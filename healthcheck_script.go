package main

import (
	"os/exec"
	"fmt"
	"os"
)

type HealthCheck_script struct {
	HealthCheckBase
	Script  string
}

func (hc HealthCheck_script) compile_config() {
	hc.HealthCheckBase.compile_config()
	logger.Info.Printf("       Specific parameters:")
	logger.Info.Printf("       - Script: '%s'", hc.Script)
}

func (hc HealthCheck_script) schedule() {
	if err := exec.Command(hc.Script).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}