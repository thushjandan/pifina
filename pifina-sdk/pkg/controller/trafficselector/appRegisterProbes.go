package trafficselector

import "github.com/thushjandan/pifina/pkg/model"

func (ts *TrafficSelector) AddAppRegisterProbe(newItem string) error {
	if tblId := ts.driver.GetTableIdByName(newItem); tblId == 0 {
		return &model.ErrNameNotFound{Msg: "Register not found", Entity: newItem}
	}

	ts.appRegisterProbesLock.Lock()
	defer ts.appRegisterProbesLock.Unlock()

	for i := range ts.appRegisterProbes {
		// if entry already exists, just ignore and exit already here.
		if ts.appRegisterProbes[i] == newItem {
			return nil
		}
	}

	ts.appRegisterProbes = append(ts.appRegisterProbes, newItem)

	return nil
}

func (ts *TrafficSelector) GetAllAppRegistersOnDevice() []string {
	return ts.driver.GetAllRegisterNames()
}

func (ts *TrafficSelector) GetAppRegisterProbes() []string {
	ts.appRegisterProbesLock.RLock()
	defer ts.appRegisterProbesLock.RUnlock()

	var copyAppRegisterProbes []string
	copyAppRegisterProbes = append(copyAppRegisterProbes, ts.appRegisterProbes...)

	return copyAppRegisterProbes
}

func (ts *TrafficSelector) RemoveAppRegisterProbe(itemToRemove string) {
	ts.appRegisterProbesLock.Lock()
	defer ts.appRegisterProbesLock.Unlock()

	newAppRegisterProbes := make([]string, 0)
	for i := range ts.appRegisterProbes {
		if ts.appRegisterProbes[i] != itemToRemove {
			newAppRegisterProbes = append(newAppRegisterProbes, ts.appRegisterProbes[i])
		}
	}

	ts.appRegisterProbes = newAppRegisterProbes
}
