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

func pfctlCmd(args []string) (*bufio.Scanner, error) {
	args = append([]string{"-q"}, args...)
	cmd := exec.Command("/sbin/pfctl", args...)
	out, err := cmd.CombinedOutput()

	outStr := string(out)

	if err != nil {
		return nil, pfctlError{err.Error() + "\n" + outStr}
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
		line := strings.Trim(out.Text(), " ")
		if ipAddress := net.ParseIP(line); ipAddress != nil {
			ret = append(ret, ipAddress)
		}
	}

	return ret, nil
}

func pfctlDelFromTable(table string, ipAddresses []net.IP) error {
	if len(ipAddresses) == 0 {
		return nil
	}
	logger.Debug.Printf("deleting from %s: %s", table, ipAddresses)
	cmd := []string{"-t", table, "-T", "del"}
	for _, ipAddress := range ipAddresses {
		cmd = append(cmd, ipAddress.String())
	}
	_, err := pfctlCmd(cmd)
	return err
}

func pfctlAddToTable(table string, ipAddresses []net.IP) error {
	if len(ipAddresses) == 0 {
		return nil
	}
	logger.Debug.Printf("adding to %s: %s", table, ipAddresses)
	cmd := []string{"-t", table, "-T", "add"}
	for _, ipAddress := range ipAddresses {
		cmd = append(cmd, ipAddress.String())
	}
	_, err := pfctlCmd(cmd)
	return err
}

func pfctlSyncTable(table string, wantSet []net.IP) error {

	var addSet, delSet []net.IP

	curSet, err := pfctlGetTable(table)
	if err != nil {
		return err
	}

	logger.Debug.Printf("have in %s: %s", table, curSet)
	logger.Debug.Printf("want in %s: %s", table, wantSet)

	// Add wanted nodes.
	for _, want := range wantSet {
		found := false
		for _, cur := range curSet {
			if want.Equal(cur) {
				found = true
			}
		}
		if found == false {
			addSet = append(addSet, want)
		}
	}

	// Remove unwanted nodes.
	for _, cur := range curSet {
		found := false
		for _, want := range wantSet {
			if cur.Equal(want) {
				found = true
			}
		}
		if found == false {
			delSet = append(delSet, cur)
		}
	}

	if err := pfctlAddToTable(table, addSet); err != nil {
		return err
	}

	if err := pfctlDelFromTable(table, delSet); err != nil {
		return err
	}

	return err
}
