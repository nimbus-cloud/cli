// This file was generated by counterfeiter
package v2fakes

import (
	"sync"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/v2"
)

type FakeCreateUserActor struct {
	NewUserStub        func(username string, password string, origin string) (v2action.User, v2action.Warnings, error)
	newUserMutex       sync.RWMutex
	newUserArgsForCall []struct {
		username string
		password string
		origin   string
	}
	newUserReturns struct {
		result1 v2action.User
		result2 v2action.Warnings
		result3 error
	}
	newUserReturnsOnCall map[int]struct {
		result1 v2action.User
		result2 v2action.Warnings
		result3 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCreateUserActor) NewUser(username string, password string, origin string) (v2action.User, v2action.Warnings, error) {
	fake.newUserMutex.Lock()
	ret, specificReturn := fake.newUserReturnsOnCall[len(fake.newUserArgsForCall)]
	fake.newUserArgsForCall = append(fake.newUserArgsForCall, struct {
		username string
		password string
		origin   string
	}{username, password, origin})
	fake.recordInvocation("NewUser", []interface{}{username, password, origin})
	fake.newUserMutex.Unlock()
	if fake.NewUserStub != nil {
		return fake.NewUserStub(username, password, origin)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fake.newUserReturns.result1, fake.newUserReturns.result2, fake.newUserReturns.result3
}

func (fake *FakeCreateUserActor) NewUserCallCount() int {
	fake.newUserMutex.RLock()
	defer fake.newUserMutex.RUnlock()
	return len(fake.newUserArgsForCall)
}

func (fake *FakeCreateUserActor) NewUserArgsForCall(i int) (string, string, string) {
	fake.newUserMutex.RLock()
	defer fake.newUserMutex.RUnlock()
	return fake.newUserArgsForCall[i].username, fake.newUserArgsForCall[i].password, fake.newUserArgsForCall[i].origin
}

func (fake *FakeCreateUserActor) NewUserReturns(result1 v2action.User, result2 v2action.Warnings, result3 error) {
	fake.NewUserStub = nil
	fake.newUserReturns = struct {
		result1 v2action.User
		result2 v2action.Warnings
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeCreateUserActor) NewUserReturnsOnCall(i int, result1 v2action.User, result2 v2action.Warnings, result3 error) {
	fake.NewUserStub = nil
	if fake.newUserReturnsOnCall == nil {
		fake.newUserReturnsOnCall = make(map[int]struct {
			result1 v2action.User
			result2 v2action.Warnings
			result3 error
		})
	}
	fake.newUserReturnsOnCall[i] = struct {
		result1 v2action.User
		result2 v2action.Warnings
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeCreateUserActor) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.newUserMutex.RLock()
	defer fake.newUserMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeCreateUserActor) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ v2.CreateUserActor = new(FakeCreateUserActor)
