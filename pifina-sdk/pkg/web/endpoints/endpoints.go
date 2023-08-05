package endpoints

import (
	"net"
	"sync"
)

type void struct{}

var member void

type PifinaEndpoint struct {
	Name     string `json:"name"`
	HostType string `json:"hostType"`
	Address  net.IP `json:"address"`
	Port     int    `json:"port"`
	GroupId  uint32 `json:"groupId"`
}

type PifinaEndpointDirectory struct {
	endpoints                map[string]*PifinaEndpoint
	defaultControllerApiPort int
	lock                     sync.RWMutex
}

func NewPifinaEndpointDirectory(port int) *PifinaEndpointDirectory {
	return &PifinaEndpointDirectory{
		endpoints:                make(map[string]*PifinaEndpoint),
		defaultControllerApiPort: port,
	}
}

func (e *PifinaEndpointDirectory) Set(newEndpoint string, hostType string, groupId uint32, address net.IP) {
	if _, ok := e.endpoints[newEndpoint]; !ok {
		e.lock.Lock()
		e.endpoints[newEndpoint] = &PifinaEndpoint{
			Name:     newEndpoint,
			HostType: hostType,
			GroupId:  groupId,
			Address:  address,
			Port:     e.defaultControllerApiPort,
		}
		e.lock.Unlock()
	}
}

func (e *PifinaEndpointDirectory) Update(endpoint string, address net.IP, port int) bool {
	e.lock.Lock()
	defer e.lock.Unlock()

	entry, ok := e.endpoints[endpoint]
	if !ok {
		return ok
	}

	entry.Address = address
	entry.Port = port

	return ok
}

func (e *PifinaEndpointDirectory) GetAll() []*PifinaEndpoint {
	endpointList := make([]*PifinaEndpoint, 0)
	e.lock.RLock()
	for key := range e.endpoints {
		endpointList = append(endpointList, e.endpoints[key])
	}
	e.lock.RUnlock()

	return endpointList
}

func (e *PifinaEndpointDirectory) Get(endpoint string) *PifinaEndpoint {
	e.lock.RLock()
	defer e.lock.RUnlock()
	result, _ := e.endpoints[endpoint]
	return result
}
