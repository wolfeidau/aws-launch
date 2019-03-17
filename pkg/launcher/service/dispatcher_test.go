package service

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/wolfeidau/aws-launch/mocks"
	"github.com/wolfeidau/aws-launch/pkg/launcher"
)

func Test_New(t *testing.T) {

	got := New(nil)
	require.NotNil(t, got)
}

func TestDispatcher_Define_Validation_Error(t *testing.T) {
	dispatcher := &Dispatcher{}
	dp := &launcher.DefineTaskParams{}
	got, err := dispatcher.DefineTask(dp)
	require.Equal(t, launcher.ErrMissingParams, err)
	require.Nil(t, got)
}

func TestDispatcher_Define_With_ECS(t *testing.T) {

	ecsLauncherMock := &mocks.Launcher{}

	ecsLauncherMock.On("DefineTask", mock.AnythingOfType("*launcher.DefineTaskParams")).Return(&launcher.DefineTaskResult{
		ID:                     "test-command:123",
		CloudwatchLogGroupName: "/aws/fargate/test-command",
		CloudwatchStreamPrefix: "ecs",
	}, nil)

	dispatcher := &Dispatcher{
		ECS: ecsLauncherMock,
	}

	dp := &launcher.DefineTaskParams{
		ECS: &launcher.ECSDefineTaskParams{
			DefinitionName: "test-command",
		},
		Tags: map[string]string{
			"TestTag": "test",
		},
		Environment: map[string]string{
			"TestEnvironment": "test",
		},
	}

	want := "test-command:123"

	got, err := dispatcher.DefineTask(dp)
	require.Nil(t, err)
	require.Equal(t, want, got.ID)
}
