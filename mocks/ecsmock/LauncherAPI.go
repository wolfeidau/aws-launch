// Code generated by mockery v1.0.0. DO NOT EDIT.

package ecsmock

import ecs "github.com/wolfeidau/aws-launch/pkg/launcher/ecs"
import mock "github.com/stretchr/testify/mock"

// LauncherAPI is an autogenerated mock type for the LauncherAPI type
type LauncherAPI struct {
	mock.Mock
}

// CleanupTask provides a mock function with given fields: _a0
func (_m *LauncherAPI) CleanupTask(_a0 *ecs.CleanupTaskParams) (*ecs.CleanupTaskResult, error) {
	ret := _m.Called(_a0)

	var r0 *ecs.CleanupTaskResult
	if rf, ok := ret.Get(0).(func(*ecs.CleanupTaskParams) *ecs.CleanupTaskResult); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ecs.CleanupTaskResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*ecs.CleanupTaskParams) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DefineTask provides a mock function with given fields: _a0
func (_m *LauncherAPI) DefineTask(_a0 *ecs.DefineTaskParams) (*ecs.DefineTaskResult, error) {
	ret := _m.Called(_a0)

	var r0 *ecs.DefineTaskResult
	if rf, ok := ret.Get(0).(func(*ecs.DefineTaskParams) *ecs.DefineTaskResult); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ecs.DefineTaskResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*ecs.DefineTaskParams) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTaskLogs provides a mock function with given fields: _a0
func (_m *LauncherAPI) GetTaskLogs(_a0 *ecs.GetTaskLogsParams) (*ecs.GetTaskLogsResult, error) {
	ret := _m.Called(_a0)

	var r0 *ecs.GetTaskLogsResult
	if rf, ok := ret.Get(0).(func(*ecs.GetTaskLogsParams) *ecs.GetTaskLogsResult); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ecs.GetTaskLogsResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*ecs.GetTaskLogsParams) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTaskStatus provides a mock function with given fields: _a0
func (_m *LauncherAPI) GetTaskStatus(_a0 *ecs.GetTaskStatusParams) (*ecs.GetTaskStatusResult, error) {
	ret := _m.Called(_a0)

	var r0 *ecs.GetTaskStatusResult
	if rf, ok := ret.Get(0).(func(*ecs.GetTaskStatusParams) *ecs.GetTaskStatusResult); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ecs.GetTaskStatusResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*ecs.GetTaskStatusParams) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LaunchTask provides a mock function with given fields: _a0
func (_m *LauncherAPI) LaunchTask(_a0 *ecs.LaunchTaskParams) (*ecs.LaunchTaskResult, error) {
	ret := _m.Called(_a0)

	var r0 *ecs.LaunchTaskResult
	if rf, ok := ret.Get(0).(func(*ecs.LaunchTaskParams) *ecs.LaunchTaskResult); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ecs.LaunchTaskResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*ecs.LaunchTaskParams) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StopTask provides a mock function with given fields: _a0
func (_m *LauncherAPI) StopTask(_a0 *ecs.StopTaskParams) (*ecs.StopTaskResult, error) {
	ret := _m.Called(_a0)

	var r0 *ecs.StopTaskResult
	if rf, ok := ret.Get(0).(func(*ecs.StopTaskParams) *ecs.StopTaskResult); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ecs.StopTaskResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*ecs.StopTaskParams) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WaitForTask provides a mock function with given fields: _a0
func (_m *LauncherAPI) WaitForTask(_a0 *ecs.WaitForTaskParams) (*ecs.WaitForTaskResult, error) {
	ret := _m.Called(_a0)

	var r0 *ecs.WaitForTaskResult
	if rf, ok := ret.Get(0).(func(*ecs.WaitForTaskParams) *ecs.WaitForTaskResult); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ecs.WaitForTaskResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*ecs.WaitForTaskParams) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
