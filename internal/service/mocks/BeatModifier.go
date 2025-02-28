// Code generated by mockery v2.50.0. DO NOT EDIT.

package mocks

import (
	context "context"

	generated "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/db/generated"
	mock "github.com/stretchr/testify/mock"

	model "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/domain/model"

	uuid "github.com/google/uuid"
)

// BeatModifier is an autogenerated mock type for the BeatModifier type
type BeatModifier struct {
	mock.Mock
}

// DeleteBeat provides a mock function with given fields: ctx, id
func (_m *BeatModifier) DeleteBeat(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for DeleteBeat")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveBeat provides a mock function with given fields: ctx, _a1
func (_m *BeatModifier) SaveBeat(ctx context.Context, _a1 model.SaveBeat) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for SaveBeat")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.SaveBeat) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveOwner provides a mock function with given fields: ctx, owner
func (_m *BeatModifier) SaveOwner(ctx context.Context, owner generated.SaveOwnerParams) error {
	ret := _m.Called(ctx, owner)

	if len(ret) == 0 {
		panic("no return value specified for SaveOwner")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, generated.SaveOwnerParams) error); ok {
		r0 = rf(ctx, owner)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateBeat provides a mock function with given fields: ctx, _a1
func (_m *BeatModifier) UpdateBeat(ctx context.Context, _a1 model.UpdateBeat) (*generated.Beat, error) {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for UpdateBeat")
	}

	var r0 *generated.Beat
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, model.UpdateBeat) (*generated.Beat, error)); ok {
		return rf(ctx, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.UpdateBeat) *generated.Beat); ok {
		r0 = rf(ctx, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*generated.Beat)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.UpdateBeat) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewBeatModifier creates a new instance of BeatModifier. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBeatModifier(t interface {
	mock.TestingT
	Cleanup(func())
}) *BeatModifier {
	mock := &BeatModifier{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
