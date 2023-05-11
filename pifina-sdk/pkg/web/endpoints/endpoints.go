package endpoints

import "sync"

type void struct{}

var member void

type PifinaEndpointDirectory struct {
	endpoints map[string]struct{}
	lock      sync.RWMutex
}

func NewPifinaEndpointDirectory() *PifinaEndpointDirectory {
	return &PifinaEndpointDirectory{
		endpoints: make(map[string]struct{}),
	}
}

func (e *PifinaEndpointDirectory) Set(newEndpoint string) {
	if _, ok := e.endpoints[newEndpoint]; !ok {
		e.lock.Lock()
		e.endpoints[newEndpoint] = member
		e.lock.Unlock()
	}
}

func (e *PifinaEndpointDirectory) GetAll() []string {
	endpointList := make([]string, 0)
	e.lock.RLock()
	for key := range e.endpoints {
		endpointList = append(endpointList, key)
	}
	e.lock.RUnlock()

	return endpointList
}
