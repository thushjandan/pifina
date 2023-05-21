package trafficselector

import "github.com/thushjandan/pifina/pkg/model"

func (ts *TrafficSelector) AddAppRegisterProbe(newItem *model.AppRegister) error {
	if tblId := ts.driver.GetTableIdByName(newItem.Name); tblId == 0 {
		return &model.ErrNameNotFound{Msg: "Register not found", Entity: newItem.Name}
	}

	ts.appRegisterProbesLock.Lock()
	defer ts.appRegisterProbesLock.Unlock()

	for i := range ts.appRegisterProbes {
		// if entry already exists, just ignore and exit already here.
		if ts.appRegisterProbes[i].Name == newItem.Name && ts.appRegisterProbes[i].Index == newItem.Index {
			return nil
		}
	}

	ts.appRegisterProbes = append(ts.appRegisterProbes, newItem)

	return nil
}

func (ts *TrafficSelector) GetAllAppRegistersOnDevice() []string {
	return ts.driver.GetAllRegisterNames()
}

func (ts *TrafficSelector) GetAppRegisterProbes() []*model.AppRegister {
	ts.appRegisterProbesLock.RLock()
	defer ts.appRegisterProbesLock.RUnlock()

	var copyAppRegisterProbes []*model.AppRegister
	if len(ts.appRegisterProbes) > 0 {
		copyAppRegisterProbes = append(copyAppRegisterProbes, ts.appRegisterProbes...)
	} else {
		copyAppRegisterProbes = make([]*model.AppRegister, 0)
	}

	return copyAppRegisterProbes
}

func (ts *TrafficSelector) RemoveAppRegisterProbe(itemToRemove *model.AppRegister) {
	ts.appRegisterProbesLock.Lock()
	defer ts.appRegisterProbesLock.Unlock()

	newAppRegisterProbes := make([]*model.AppRegister, 0)
	for i := range ts.appRegisterProbes {
		if ts.appRegisterProbes[i].Name != itemToRemove.Name && ts.appRegisterProbes[i].Index != itemToRemove.Index {
			newAppRegisterProbes = append(newAppRegisterProbes, ts.appRegisterProbes[i])
		}
	}

	ts.appRegisterProbes = newAppRegisterProbes
}
