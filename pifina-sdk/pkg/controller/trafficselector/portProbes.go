package trafficselector

func (ts *TrafficSelector) GetAllAvailablePorts() []string {
	return ts.driver.GetAvailablePortNames()
}

func (ts *TrafficSelector) AddPortToMonitor(newItem string) {
	ts.monitoredDevPortsLock.Lock()
	defer ts.monitoredDevPortsLock.Unlock()

	for i := range ts.monitoredDevPorts {
		// if entry already exists, just ignore and exit already here.
		if ts.monitoredDevPorts[i] == newItem {
			return
		}
	}

	ts.monitoredDevPorts = append(ts.monitoredDevPorts, newItem)
}

func (ts *TrafficSelector) GetMonitoredPorts() []string {
	ts.monitoredDevPortsLock.RLock()
	defer ts.monitoredDevPortsLock.RUnlock()

	var copyMonitoredPorts []string
	if len(ts.monitoredDevPorts) > 0 {
		copyMonitoredPorts = append(copyMonitoredPorts, ts.monitoredDevPorts...)
	} else {
		copyMonitoredPorts = make([]string, 0)
	}

	return copyMonitoredPorts
}

func (ts *TrafficSelector) RemovePortToMonitor(itemToRemove string) {
	ts.monitoredDevPortsLock.Lock()
	defer ts.monitoredDevPortsLock.Unlock()

	newPortsToMonitor := make([]string, 0)
	for i := range ts.monitoredDevPorts {
		if ts.monitoredDevPorts[i] != itemToRemove {
			newPortsToMonitor = append(newPortsToMonitor, ts.monitoredDevPorts[i])
		}
	}

	ts.monitoredDevPorts = newPortsToMonitor
}
