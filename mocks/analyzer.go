// Code generated by mockery v2.33.1. DO NOT EDIT.

package mocks

import (
	context "context"

	entity "github.com/fadyat/i4u/internal/entity"
	mock "github.com/stretchr/testify/mock"
)

// Analyzer is an autogenerated mock type for the Analyzer type
type Analyzer struct {
	mock.Mock
}

// IsInternshipRequest provides a mock function with given fields: _a0, _a1
func (_m *Analyzer) IsInternshipRequest(_a0 context.Context, _a1 entity.Message) (bool, error) {
	ret := _m.Called(_a0, _a1)

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, entity.Message) (bool, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, entity.Message) bool); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, entity.Message) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewAnalyzer creates a new instance of Analyzer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAnalyzer(t interface {
	mock.TestingT
	Cleanup(func())
}) *Analyzer {
	mock := &Analyzer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
