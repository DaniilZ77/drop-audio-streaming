// Code generated by mockery v2.50.0. DO NOT EDIT.

package mocks

import (
	context "context"

	generated "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/db/generated"
	mock "github.com/stretchr/testify/mock"

	model "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/domain/model"

	uuid "github.com/google/uuid"
)

// BeatProvider is an autogenerated mock type for the BeatProvider type
type BeatProvider struct {
	mock.Mock
}

// GetBeatByID provides a mock function with given fields: ctx, id
func (_m *BeatProvider) GetBeatByID(ctx context.Context, id uuid.UUID) (*generated.Beat, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetBeatByID")
	}

	var r0 *generated.Beat
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*generated.Beat, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *generated.Beat); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*generated.Beat)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBeatParams provides a mock function with given fields: ctx
func (_m *BeatProvider) GetBeatParams(ctx context.Context) (*model.BeatAttributes, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetBeatParams")
	}

	var r0 *model.BeatAttributes
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*model.BeatAttributes, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *model.BeatAttributes); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.BeatAttributes)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBeats provides a mock function with given fields: ctx, params
func (_m *BeatProvider) GetBeats(ctx context.Context, params model.GetBeatsParams) ([]model.Beat, *uint64, error) {
	ret := _m.Called(ctx, params)

	if len(ret) == 0 {
		panic("no return value specified for GetBeats")
	}

	var r0 []model.Beat
	var r1 *uint64
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, model.GetBeatsParams) ([]model.Beat, *uint64, error)); ok {
		return rf(ctx, params)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.GetBeatsParams) []model.Beat); ok {
		r0 = rf(ctx, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.Beat)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.GetBeatsParams) *uint64); ok {
		r1 = rf(ctx, params)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*uint64)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, model.GetBeatsParams) error); ok {
		r2 = rf(ctx, params)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetOwnerByBeatID provides a mock function with given fields: ctx, beatID
func (_m *BeatProvider) GetOwnerByBeatID(ctx context.Context, beatID uuid.UUID) (*generated.BeatsOwner, error) {
	ret := _m.Called(ctx, beatID)

	if len(ret) == 0 {
		panic("no return value specified for GetOwnerByBeatID")
	}

	var r0 *generated.BeatsOwner
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*generated.BeatsOwner, error)); ok {
		return rf(ctx, beatID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *generated.BeatsOwner); ok {
		r0 = rf(ctx, beatID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*generated.BeatsOwner)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, beatID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewBeatProvider creates a new instance of BeatProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBeatProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *BeatProvider {
	mock := &BeatProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
