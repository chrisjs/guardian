// This file was generated by counterfeiter
package fakes

import (
	"sync"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/guardian/gardener"
)

type FakePropertyManager struct {
	AllStub        func(handle string) (props garden.Properties, err error)
	allMutex       sync.RWMutex
	allArgsForCall []struct {
		handle string
	}
	allReturns struct {
		result1 garden.Properties
		result2 error
	}
	SetStub        func(handle string, name string, value string) error
	setMutex       sync.RWMutex
	setArgsForCall []struct {
		handle string
		name   string
		value  string
	}
	setReturns struct {
		result1 error
	}
	RemoveStub        func(handle string, name string) error
	removeMutex       sync.RWMutex
	removeArgsForCall []struct {
		handle string
		name   string
	}
	removeReturns struct {
		result1 error
	}
	GetStub        func(handle string, name string) (string, error)
	getMutex       sync.RWMutex
	getArgsForCall []struct {
		handle string
		name   string
	}
	getReturns struct {
		result1 string
		result2 error
	}
	CreateKeySpaceStub        func(string) error
	createKeySpaceMutex       sync.RWMutex
	createKeySpaceArgsForCall []struct {
		arg1 string
	}
	createKeySpaceReturns struct {
		result1 error
	}
	DestroyKeySpaceStub        func(string) error
	destroyKeySpaceMutex       sync.RWMutex
	destroyKeySpaceArgsForCall []struct {
		arg1 string
	}
	destroyKeySpaceReturns struct {
		result1 error
	}
}

func (fake *FakePropertyManager) All(handle string) (props garden.Properties, err error) {
	fake.allMutex.Lock()
	fake.allArgsForCall = append(fake.allArgsForCall, struct {
		handle string
	}{handle})
	fake.allMutex.Unlock()
	if fake.AllStub != nil {
		return fake.AllStub(handle)
	} else {
		return fake.allReturns.result1, fake.allReturns.result2
	}
}

func (fake *FakePropertyManager) AllCallCount() int {
	fake.allMutex.RLock()
	defer fake.allMutex.RUnlock()
	return len(fake.allArgsForCall)
}

func (fake *FakePropertyManager) AllArgsForCall(i int) string {
	fake.allMutex.RLock()
	defer fake.allMutex.RUnlock()
	return fake.allArgsForCall[i].handle
}

func (fake *FakePropertyManager) AllReturns(result1 garden.Properties, result2 error) {
	fake.AllStub = nil
	fake.allReturns = struct {
		result1 garden.Properties
		result2 error
	}{result1, result2}
}

func (fake *FakePropertyManager) Set(handle string, name string, value string) error {
	fake.setMutex.Lock()
	fake.setArgsForCall = append(fake.setArgsForCall, struct {
		handle string
		name   string
		value  string
	}{handle, name, value})
	fake.setMutex.Unlock()
	if fake.SetStub != nil {
		return fake.SetStub(handle, name, value)
	} else {
		return fake.setReturns.result1
	}
}

func (fake *FakePropertyManager) SetCallCount() int {
	fake.setMutex.RLock()
	defer fake.setMutex.RUnlock()
	return len(fake.setArgsForCall)
}

func (fake *FakePropertyManager) SetArgsForCall(i int) (string, string, string) {
	fake.setMutex.RLock()
	defer fake.setMutex.RUnlock()
	return fake.setArgsForCall[i].handle, fake.setArgsForCall[i].name, fake.setArgsForCall[i].value
}

func (fake *FakePropertyManager) SetReturns(result1 error) {
	fake.SetStub = nil
	fake.setReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakePropertyManager) Remove(handle string, name string) error {
	fake.removeMutex.Lock()
	fake.removeArgsForCall = append(fake.removeArgsForCall, struct {
		handle string
		name   string
	}{handle, name})
	fake.removeMutex.Unlock()
	if fake.RemoveStub != nil {
		return fake.RemoveStub(handle, name)
	} else {
		return fake.removeReturns.result1
	}
}

func (fake *FakePropertyManager) RemoveCallCount() int {
	fake.removeMutex.RLock()
	defer fake.removeMutex.RUnlock()
	return len(fake.removeArgsForCall)
}

func (fake *FakePropertyManager) RemoveArgsForCall(i int) (string, string) {
	fake.removeMutex.RLock()
	defer fake.removeMutex.RUnlock()
	return fake.removeArgsForCall[i].handle, fake.removeArgsForCall[i].name
}

func (fake *FakePropertyManager) RemoveReturns(result1 error) {
	fake.RemoveStub = nil
	fake.removeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakePropertyManager) Get(handle string, name string) (string, error) {
	fake.getMutex.Lock()
	fake.getArgsForCall = append(fake.getArgsForCall, struct {
		handle string
		name   string
	}{handle, name})
	fake.getMutex.Unlock()
	if fake.GetStub != nil {
		return fake.GetStub(handle, name)
	} else {
		return fake.getReturns.result1, fake.getReturns.result2
	}
}

func (fake *FakePropertyManager) GetCallCount() int {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	return len(fake.getArgsForCall)
}

func (fake *FakePropertyManager) GetArgsForCall(i int) (string, string) {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	return fake.getArgsForCall[i].handle, fake.getArgsForCall[i].name
}

func (fake *FakePropertyManager) GetReturns(result1 string, result2 error) {
	fake.GetStub = nil
	fake.getReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakePropertyManager) CreateKeySpace(arg1 string) error {
	fake.createKeySpaceMutex.Lock()
	fake.createKeySpaceArgsForCall = append(fake.createKeySpaceArgsForCall, struct {
		arg1 string
	}{arg1})
	fake.createKeySpaceMutex.Unlock()
	if fake.CreateKeySpaceStub != nil {
		return fake.CreateKeySpaceStub(arg1)
	} else {
		return fake.createKeySpaceReturns.result1
	}
}

func (fake *FakePropertyManager) CreateKeySpaceCallCount() int {
	fake.createKeySpaceMutex.RLock()
	defer fake.createKeySpaceMutex.RUnlock()
	return len(fake.createKeySpaceArgsForCall)
}

func (fake *FakePropertyManager) CreateKeySpaceArgsForCall(i int) string {
	fake.createKeySpaceMutex.RLock()
	defer fake.createKeySpaceMutex.RUnlock()
	return fake.createKeySpaceArgsForCall[i].arg1
}

func (fake *FakePropertyManager) CreateKeySpaceReturns(result1 error) {
	fake.CreateKeySpaceStub = nil
	fake.createKeySpaceReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakePropertyManager) DestroyKeySpace(arg1 string) error {
	fake.destroyKeySpaceMutex.Lock()
	fake.destroyKeySpaceArgsForCall = append(fake.destroyKeySpaceArgsForCall, struct {
		arg1 string
	}{arg1})
	fake.destroyKeySpaceMutex.Unlock()
	if fake.DestroyKeySpaceStub != nil {
		return fake.DestroyKeySpaceStub(arg1)
	} else {
		return fake.destroyKeySpaceReturns.result1
	}
}

func (fake *FakePropertyManager) DestroyKeySpaceCallCount() int {
	fake.destroyKeySpaceMutex.RLock()
	defer fake.destroyKeySpaceMutex.RUnlock()
	return len(fake.destroyKeySpaceArgsForCall)
}

func (fake *FakePropertyManager) DestroyKeySpaceArgsForCall(i int) string {
	fake.destroyKeySpaceMutex.RLock()
	defer fake.destroyKeySpaceMutex.RUnlock()
	return fake.destroyKeySpaceArgsForCall[i].arg1
}

func (fake *FakePropertyManager) DestroyKeySpaceReturns(result1 error) {
	fake.DestroyKeySpaceStub = nil
	fake.destroyKeySpaceReturns = struct {
		result1 error
	}{result1}
}

var _ gardener.PropertyManager = new(FakePropertyManager)
