package neohost

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/model"
)

type NeoHostDriver struct {
	logger  hclog.Logger
	sdkPath string
	neoMode string
	neoPort int
}

type NeoHostDriverOptions struct {
	Logger  hclog.Logger
	SDKPath string
	NEOMode string
	NEOPort int
}

func NewNeoHostDriver(options *NeoHostDriverOptions) *NeoHostDriver {
	return &NeoHostDriver{
		logger:  options.Logger,
		sdkPath: options.SDKPath,
		neoMode: options.NEOMode,
		neoPort: options.NEOPort,
	}
}

func (d *NeoHostDriver) ListMlxNetworkCards() (*model.NeoHostDeviceList, error) {
	pythonExecPath, err := exec.LookPath("python3")
	if err != nil {
		return nil, err
	}

	// Check if sdk exists
	if _, err := os.Stat(d.sdkPath); errors.Is(err, os.ErrNotExist) {
		d.logger.Error("NEO-Host SDK has not been found. Install NEO-Host SDK first or provide correct path to the SDK", "path", d.sdkPath, "err", err)
		return nil, err
	}

	progPath := filepath.Join(d.sdkPath, "get_system_devices.py")

	// Check if progPath exists
	if _, err := os.Stat(progPath); errors.Is(err, os.ErrNotExist) {
		d.logger.Error("get_system_devices.py has not been found in NEO-Host SDK. Maybe reinstall NEO-Host SDK", "path", progPath, "err", err)
		return nil, err
	}

	var cmd *exec.Cmd
	if d.neoPort == 0 {
		cmd = exec.Command(pythonExecPath, progPath, d.neoMode, "--output=JSON")
	} else {
		cmd = exec.Command(pythonExecPath, progPath, d.neoMode, "--output=JSON", fmt.Sprintf("--port=%d", d.neoPort))
	}
	d.logger.Debug("Retrieving system devices", "cmd", cmd.Args)

	var commandResult *model.NeoHostDeviceList

	stdout, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if stdout == nil {
				d.logger.Error("Error occured during retrieving Mellanox NICs information", "err", exitErr.Stderr, "command", cmd.Args)
			} else {
				jsonErr := json.Unmarshal(stdout, &commandResult)
				if jsonErr != nil {
					d.logger.Error("Error occured during retrieving Mellanox NICs information", "err", string(stdout), "command", cmd.Args)
					return nil, &model.ErrNotReady{Msg: "NEO SDK returned an error. See previous log message."}
				}
				d.logger.Error("NEO Host SDK returned an error during retrieving Mellanox NICs information.", "err", commandResult.Error.Message, "command", cmd.Args)
			}
		}
		return nil, &model.ErrNotReady{Msg: "NEO SDK returned an error. See previous log message."}
	}

	err = json.Unmarshal(stdout, &commandResult)

	return commandResult, err
}

func (d *NeoHostDriver) GetPerformanceCounters(devUid string) (*model.NeoHostPerfCounterResult, error) {
	pythonExecPath, err := exec.LookPath("python3")
	if err != nil {
		return nil, err
	}

	// Check if sdk exists
	if _, err := os.Stat(d.sdkPath); errors.Is(err, os.ErrNotExist) {
		d.logger.Error("NEO-Host SDK has not been found. Install NEO-Host SDK first or provide correct path to the SDK", "path", d.sdkPath, "err", err)
		return nil, err
	}

	progPath := filepath.Join(d.sdkPath, "get_device_performance_counters.py")

	// Check if progPath exists
	if _, err := os.Stat(progPath); errors.Is(err, os.ErrNotExist) {
		d.logger.Error("get_device_performance_counters.py has not been found in NEO-Host SDK. Maybe reinstall NEO-Host SDK", "path", progPath, "err", err)
		return nil, err
	}

	var cmd *exec.Cmd
	if d.neoPort == 0 {
		cmd = exec.Command(pythonExecPath, progPath, d.neoMode, fmt.Sprintf("--dev-uid=%s", devUid), "--output=JSON")
	} else {
		cmd = exec.Command(pythonExecPath, progPath, d.neoMode, fmt.Sprintf("--dev-uid=%s", devUid), "--output=JSON", fmt.Sprintf("--port=%d", d.neoPort))
	}
	d.logger.Debug("Retrieving performance counters devices", "cmd", cmd.Args)

	var commandResult *model.NeoHostPerfCounterResult

	stdout, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if stdout == nil {
				d.logger.Error("Error occured during retrieving performance counters from NIC", "err", exitErr.Stderr)
			} else {
				d.logger.Error("NEO Host SDK returned an error during retrieving performance counters from NIC", "command", cmd.Args, "err", string(stdout))
			}
			return nil, &model.ErrNotReady{Msg: "NEO SDK returned an error. See previous log message."}
		}
		return nil, err
	}

	err = json.Unmarshal(stdout, &commandResult)

	return commandResult, err
}
