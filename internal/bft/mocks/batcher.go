// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Batcher is an autogenerated mock type for the Batcher type
type Batcher struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Batcher) Close() {
	_m.Called()
}

// Closed provides a mock function with given fields:
func (_m *Batcher) Closed() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// NextBatch provides a mock function with given fields:
func (_m *Batcher) NextBatch() [][]byte {
	ret := _m.Called()

	var r0 [][]byte
	if rf, ok := ret.Get(0).(func() [][]byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]byte)
		}
	}

	return r0
}

// Reset provides a mock function with given fields:
func (_m *Batcher) Reset() {
	_m.Called()
}
