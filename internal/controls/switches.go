package controls

import (
	"os/exec"
	"time"
)

type Switch struct {
	Name            string        `mapstructure:"name"`
	OnCmd           string        `mapstructure:"turn_on"`
	OffCmd          string        `mapstructure:"turn_off"`
	StateCmd        string        `mapstructure:"get_state"`
	RefreshInterval time.Duration `mapstructure:"refresh"`
}

func (sw *Switch) SwitchOnOff(state bool) (string, error) {
	if state {
		return run(sw.OnCmd)
	} else {
		return run(sw.OffCmd)
	}
}

func (sw *Switch) GetState() (bool, string, error) {
	out, err := run(sw.StateCmd)
	if exitError, ok := err.(*exec.ExitError); ok {
		if exitError.ExitCode() == 1 {
			return false, out, nil
		} else {
			return false, out, err
		}
	} else {
		return err == nil, out, err
	}
}

func run(command string) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", command)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
