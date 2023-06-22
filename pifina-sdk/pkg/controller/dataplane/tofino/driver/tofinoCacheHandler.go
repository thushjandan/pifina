package driver

import (
	"strings"
)

func (driver *TofinoDriver) createP4TableIndex() {
	driver.indexP4Tables = make(map[string]int)
	driver.indexByIdP4Tables = make(map[uint32]int)
	driver.extraProbeNameCache = make([]string, 0)
	for i := range driver.P4Tables {
		name := driver.P4Tables[i].Name
		driver.indexP4Tables[name] = i
		driver.indexByIdP4Tables[driver.P4Tables[i].Id] = i
		// Find the full table name of each probe and cache it
		for _, probe := range PROBE_TABLES {
			if strings.Contains(name, probe) {
				driver.probeTableMap[probe] = name
				break
			}
		}
		// Check if there are any extra probes installed
		if strings.Contains(name, PROBE_EXTRA_PREFIX) {
			tblNameSplit := strings.Split(name, ".")
			shortTblName := tblNameSplit[len(tblNameSplit)-1]
			// Add the short name to the table cache
			driver.probeTableMap[shortTblName] = name
			// Track extra probes separately
			driver.extraProbeNameCache = append(driver.extraProbeNameCache, shortTblName)
		}
	}
}

func (driver *TofinoDriver) createNonP4TableIndex() {
	driver.indexNonP4Tables = make(map[string]int)
	driver.indexByIdNonP4Tables = make(map[uint32]int)
	for i := range driver.NonP4Tables {
		driver.indexNonP4Tables[driver.NonP4Tables[i].Name] = i
		driver.indexByIdNonP4Tables[driver.NonP4Tables[i].Id] = i
	}
}

// Check if an item is in the list of predefined probes
func (driver *TofinoDriver) IsInProbeTable(item string) bool {
	for i := range PROBE_TABLES {
		if PROBE_TABLES[i] == item {
			return true
		}
	}
	return false
}

func (driver *TofinoDriver) GetTableIdByName(tblName string) uint32 {
	tblId := uint32(0)
	// Find table name in index
	if sliceIdx, ok := driver.indexP4Tables[tblName]; ok {
		// Table name has been found in hash table
		return driver.P4Tables[sliceIdx].Id
	}

	if sliceIdx, ok := driver.indexNonP4Tables[tblName]; ok {
		// Table name has been found in hash table
		return driver.NonP4Tables[sliceIdx].Id
	}

	return tblId
}

func (driver *TofinoDriver) GetTableNameById(tblId uint32) string {
	tblName := ""
	// Find table name in index
	if sliceIdx, ok := driver.indexByIdP4Tables[tblId]; ok {
		// Table name has been found in hash table
		return driver.P4Tables[sliceIdx].Name
	}

	// Find table name in index
	if sliceIdx, ok := driver.indexByIdNonP4Tables[tblId]; ok {
		// Table name has been found in hash table
		return driver.NonP4Tables[sliceIdx].Name
	}

	return tblName
}

// Find full table name by the short name of the table
// e.g. PF_EGRESS_START_CNT => pipe.SwitchEgress.pfEgressStartProbe.PF_EGRESS_START_CNT
func (driver *TofinoDriver) FindTableNameByShortName(shortName string) string {
	if tblName, ok := driver.probeTableMap[shortName]; ok {
		return tblName
	}
	return ""
}

func (driver *TofinoDriver) GetKeyIdByName(tblName, keyName string) uint32 {
	keyId := uint32(0)
	// Find table name in index
	if sliceIdx, ok := driver.indexP4Tables[tblName]; ok {
		// Table name has been found in hash table
		for keyIdx := range driver.P4Tables[sliceIdx].Key {
			if driver.P4Tables[sliceIdx].Key[keyIdx].Name == keyName {
				return driver.P4Tables[sliceIdx].Key[keyIdx].Id
			}
		}
	}

	if sliceIdx, ok := driver.indexNonP4Tables[tblName]; ok {
		// Table name has been found in hash table
		for keyIdx := range driver.NonP4Tables[sliceIdx].Key {
			if driver.NonP4Tables[sliceIdx].Key[keyIdx].Name == keyName {
				return driver.NonP4Tables[sliceIdx].Key[keyIdx].Id
			}
		}
	}
	return keyId
}

func (driver *TofinoDriver) GetActionIdByName(tblName, actionName string) uint32 {
	actionId := uint32(0)
	// Find table name in index
	if sliceIdx, ok := driver.indexP4Tables[tblName]; ok {
		// Table name has been found in hash table
		for actionIdx := range driver.P4Tables[sliceIdx].ActionSpecs {
			if driver.P4Tables[sliceIdx].ActionSpecs[actionIdx].Name == actionName {
				return driver.P4Tables[sliceIdx].ActionSpecs[actionIdx].Id
			}
		}
	}
	return actionId
}

func (driver *TofinoDriver) GetActionDataWidthByName(tblName, actionName string, dataName string) uint32 {
	actionDataWidth := uint32(0)
	// Find table name in index
	if sliceIdx, ok := driver.indexP4Tables[tblName]; ok {
		// Table name has been found in hash table
		for actionIdx := range driver.P4Tables[sliceIdx].ActionSpecs {
			if driver.P4Tables[sliceIdx].ActionSpecs[actionIdx].Name == actionName {
				for dataIdx := range driver.P4Tables[sliceIdx].ActionSpecs[actionIdx].Data {
					if driver.P4Tables[sliceIdx].ActionSpecs[actionIdx].Data[dataIdx].Name == dataName {
						return driver.P4Tables[sliceIdx].ActionSpecs[actionIdx].Data[dataIdx].Type.Width
					}
				}
			}
		}
	}
	return actionDataWidth
}

// Find full action name of an action.
func (driver *TofinoDriver) FindFullActionName(tblName, partialActionName string) string {
	actionName := ""
	// Find table name in index
	if sliceIdx, ok := driver.indexP4Tables[tblName]; ok {
		// Table name has been found in hash table
		for actionIdx := range driver.P4Tables[sliceIdx].ActionSpecs {
			if strings.Contains(driver.P4Tables[sliceIdx].ActionSpecs[actionIdx].Name, partialActionName) {
				return driver.P4Tables[sliceIdx].ActionSpecs[actionIdx].Name
			}
		}
	}
	return actionName
}

func (driver *TofinoDriver) GetDataIdByName(tblName, actionName, dataName string) uint32 {
	dataId := uint32(0)
	// Find table name in index
	if sliceIdx, ok := driver.indexP4Tables[tblName]; ok {
		// Table name has been found in hash table
		for actionIdx := range driver.P4Tables[sliceIdx].ActionSpecs {
			actionSpecObj := driver.P4Tables[sliceIdx].ActionSpecs[actionIdx]
			if actionSpecObj.Name == actionName {
				for dataIdx := range actionSpecObj.Data {
					if actionSpecObj.Data[dataIdx].Name == dataName {
						return actionSpecObj.Data[dataIdx].Id
					}
				}
			}
		}
	}
	return dataId
}

func (driver *TofinoDriver) GetSingletonDataIdByName(tblName, dataName string) uint32 {
	dataId := uint32(0)
	// Find table name in index
	if sliceIdx, ok := driver.indexP4Tables[tblName]; ok {
		// Table name has been found in hash table
		for dataIdx := range driver.P4Tables[sliceIdx].Data {
			dataObj := driver.P4Tables[sliceIdx].Data[dataIdx]
			if dataObj.Singleton.Name == dataName {
				return dataObj.Singleton.Id
			}
		}
	}
	return dataId
}

func (driver *TofinoDriver) GetSingletonDataNameById(tblName string, dataId uint32) string {
	dataName := ""
	// Find table name in index
	if sliceIdx, ok := driver.indexP4Tables[tblName]; ok {
		// Table name has been found in hash table
		for dataIdx := range driver.P4Tables[sliceIdx].Data {
			dataObj := driver.P4Tables[sliceIdx].Data[dataIdx]
			if dataObj.Singleton.Id == dataId {
				return dataObj.Singleton.Name
			}
		}
	}
	// Find table name in index
	if sliceIdx, ok := driver.indexNonP4Tables[tblName]; ok {
		// Table name has been found in hash table
		for dataIdx := range driver.NonP4Tables[sliceIdx].Data {
			dataObj := driver.NonP4Tables[sliceIdx].Data[dataIdx]
			if dataObj.Singleton.Id == dataId {
				return dataObj.Singleton.Name
			}
		}
	}
	return dataName
}

func (driver *TofinoDriver) GetSingletonDataIdLikeName(tblName, shortDataName string) uint32 {
	dataId := uint32(0)
	// Find table name in index
	if sliceIdx, ok := driver.indexP4Tables[tblName]; ok {
		// Table name has been found in hash table
		for dataIdx := range driver.P4Tables[sliceIdx].Data {
			dataObj := driver.P4Tables[sliceIdx].Data[dataIdx]
			if strings.Contains(dataObj.Singleton.Name, shortDataName) {
				return dataObj.Singleton.Id
			}
		}
	}
	return dataId
}

func (driver *TofinoDriver) GetExtraProbes() []string {
	// Return from cache if exists
	return driver.extraProbeNameCache
}
