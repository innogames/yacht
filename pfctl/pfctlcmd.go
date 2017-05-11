package pfctl

import (
	"bufio"
	"github.com/innogames/yacht/logger"
	"net"
	"os/exec"
	"strings"
)

type pfctlError struct {
	s string
}

func (e pfctlError) Error() string {
	return "pfctl: " + e.s
}

func pfctlCmd(arg []string) (*bufio.Scanner, error) {
	cmd := exec.Command("/sbin/pfctl", arg...)
	out, err := cmd.CombinedOutput()

	outStr := string(out)

	if err != nil {
		return nil, pfctlError{err.Error() + outStr}
	}

	scanner := bufio.NewScanner(strings.NewReader(outStr))

	return scanner, nil
}

func pfctlGetTable(table string) ([]net.IP, error) {
	var ret []net.IP

	out, err := pfctlCmd([]string{"-t", table, "-Ts"})
	if err != nil {
		return nil, err
	}

	for out.Scan() {
		line := out.Text()
		logger.Debug.Printf("got line '%s'", line)
	}

	return ret, nil
}

func pfctlSyncTable(table string, nodes []string) error {

	_, err := pfctlGetTable(table)

	return err
}
