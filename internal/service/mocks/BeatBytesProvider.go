// Code generated by mockery v2.50.0. DO NOT EDIT.

package mocks

import (
	context "context"
	io "io"

	mock "github.com/stretchr/testify/mock"
)

// BeatBytesProvider is an autogenerated mock type for the BeatBytesProvider type
type BeatBytesProvider struct {
	mock.Mock
}

// GetBeatBytes provides a mock function with given fields: ctx, path, s, e
func (_m *BeatBytesProvider) GetBeatBytes(ctx context.Context, path string, s *int, e *int) (io.ReadCloser, *int, *string, error) {
	ret := _m.Called(ctx, path, s, e)

	if len(ret) == 0 {
		panic("no return value specified for GetBeatBytes")
	}

	var r0 io.ReadCloser
	var r1 *int
	var r2 *string
	var r3 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *int, *int) (io.ReadCloser, *int, *string, error)); ok {
		return rf(ctx, path, s, e)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *int, *int) io.ReadCloser); ok {
		r0 = rf(ctx, path, s, e)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadCloser)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *int, *int) *int); ok {
		r1 = rf(ctx, path, s, e)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*int)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, string, *int, *int) *string); ok {
		r2 = rf(ctx, path, s, e)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(*string)
		}
	}

	if rf, ok := ret.Get(3).(func(context.Context, string, *int, *int) error); ok {
		r3 = rf(ctx, path, s, e)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// NewBeatBytesProvider creates a new instance of BeatBytesProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBeatBytesProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *BeatBytesProvider {
	mock := &BeatBytesProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
