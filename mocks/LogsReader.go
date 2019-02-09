// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import cwlogs "github.com/wolfeidau/aws-launch/pkg/cwlogs"
import mock "github.com/stretchr/testify/mock"

// LogsReader is an autogenerated mock type for the LogsReader type
type LogsReader struct {
	mock.Mock
}

// ReadLogs provides a mock function with given fields: _a0
func (_m *LogsReader) ReadLogs(_a0 *cwlogs.ReadLogsParams) (*cwlogs.ReadLogsResult, error) {
	ret := _m.Called(_a0)

	var r0 *cwlogs.ReadLogsResult
	if rf, ok := ret.Get(0).(func(*cwlogs.ReadLogsParams) *cwlogs.ReadLogsResult); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*cwlogs.ReadLogsResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*cwlogs.ReadLogsParams) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
