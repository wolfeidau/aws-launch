// Code generated by mockery v1.0.0. DO NOT EDIT.

package codebuildmock

import codebuild "github.com/wolfeidau/aws-launch/pkg/launcher/codebuild"
import mock "github.com/stretchr/testify/mock"

// LauncherAPI is an autogenerated mock type for the LauncherAPI type
type LauncherAPI struct {
	mock.Mock
}

// CleanupTask provides a mock function with given fields: _a0
func (_m *LauncherAPI) CleanupTask(_a0 *codebuild.CleanupTaskParams) (*codebuild.CleanupTaskResult, error) {
	ret := _m.Called(_a0)

	var r0 *codebuild.CleanupTaskResult
	if rf, ok := ret.Get(0).(func(*codebuild.CleanupTaskParams) *codebuild.CleanupTaskResult); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*codebuild.CleanupTaskResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*codebuild.CleanupTaskParams) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DefineTask provides a mock function with given fields: _a0
func (_m *LauncherAPI) DefineTask(_a0 *codebuild.DefineTaskParams) (*codebuild.DefineTaskResult, error) {
	ret := _m.Called(_a0)

	var r0 *codebuild.DefineTaskResult
	if rf, ok := ret.Get(0).(func(*codebuild.DefineTaskParams) *codebuild.DefineTaskResult); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*codebuild.DefineTaskResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*codebuild.DefineTaskParams) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTaskLogs provides a mock function with given fields: _a0
func (_m *LauncherAPI) GetTaskLogs(_a0 *codebuild.GetTaskLogsParams) (*codebuild.GetTaskLogsResult, error) {
	ret := _m.Called(_a0)

	var r0 *codebuild.GetTaskLogsResult
	if rf, ok := ret.Get(0).(func(*codebuild.GetTaskLogsParams) *codebuild.GetTaskLogsResult); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*codebuild.GetTaskLogsResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*codebuild.GetTaskLogsParams) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTaskStatus provides a mock function with given fields: _a0
func (_m *LauncherAPI) GetTaskStatus(_a0 *codebuild.GetTaskStatusParams) (*codebuild.GetTaskStatusResult, error) {
	ret := _m.Called(_a0)

	var r0 *codebuild.GetTaskStatusResult
	if rf, ok := ret.Get(0).(func(*codebuild.GetTaskStatusParams) *codebuild.GetTaskStatusResult); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*codebuild.GetTaskStatusResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*codebuild.GetTaskStatusParams) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LaunchTask provides a mock function with given fields: _a0
func (_m *LauncherAPI) LaunchTask(_a0 *codebuild.LaunchTaskParams) (*codebuild.LaunchTaskResult, error) {
	ret := _m.Called(_a0)

	var r0 *codebuild.LaunchTaskResult
	if rf, ok := ret.Get(0).(func(*codebuild.LaunchTaskParams) *codebuild.LaunchTaskResult); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*codebuild.LaunchTaskResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*codebuild.LaunchTaskParams) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StopTask provides a mock function with given fields: _a0
func (_m *LauncherAPI) StopTask(_a0 *codebuild.StopTaskParams) (*codebuild.StopTaskResult, error) {
	ret := _m.Called(_a0)

	var r0 *codebuild.StopTaskResult
	if rf, ok := ret.Get(0).(func(*codebuild.StopTaskParams) *codebuild.StopTaskResult); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*codebuild.StopTaskResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*codebuild.StopTaskParams) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WaitForTask provides a mock function with given fields: _a0
func (_m *LauncherAPI) WaitForTask(_a0 *codebuild.WaitForTaskParams) (*codebuild.WaitForTaskResult, error) {
	ret := _m.Called(_a0)

	var r0 *codebuild.WaitForTaskResult
	if rf, ok := ret.Get(0).(func(*codebuild.WaitForTaskParams) *codebuild.WaitForTaskResult); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*codebuild.WaitForTaskResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*codebuild.WaitForTaskParams) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
