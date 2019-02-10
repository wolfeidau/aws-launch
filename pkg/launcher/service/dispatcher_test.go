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

func TestDispatcher_DefineAndLaunch_Validation_Error(t *testing.T) {
	dispatcher := &Dispatcher{}
	dp := &launcher.DefineAndLaunchParams{}
	got, err := dispatcher.DefineAndLaunch(dp)
	require.Equal(t, launcher.ErrMissingParams, err)
	require.Nil(t, got)
}

func TestDispatcher_DefineAndLaunch_With_ECS(t *testing.T) {

	ecsLauncherMock := &mocks.Launcher{}

	ecsLauncherMock.On("DefineAndLaunch", mock.AnythingOfType("*launcher.DefineAndLaunchParams")).Return(&launcher.DefineAndLaunchResult{
		BaseTaskResult: &launcher.BaseTaskResult{
			ID: "arn:aws:ecs:ap-southeast-2:123456789012:task/wolfeidau-ecs-dev-Cluster-1234567890123/dece5e631c854b0d9edd5d93e91d5b8c",
		},
		DefinitionID:           "test-command:123",
		CloudwatchLogGroupName: "/aws/fargate/test-command",
		CloudwatchStreamPrefix: "ecs",
	}, nil)

	dispatcher := &Dispatcher{
		ECS: ecsLauncherMock,
	}

	dp := &launcher.DefineAndLaunchParams{
		ECS: &launcher.ECSDefineAndLaunchParams{
			ClusterName:    "abc123",
			DefinitionName: "test-command",
		},
		Tags: map[string]string{
			"TestTag": "test",
		},
		Environment: map[string]string{
			"TestEnvironment": "test",
		},
	}

	want := "arn:aws:ecs:ap-southeast-2:123456789012:task/wolfeidau-ecs-dev-Cluster-1234567890123/dece5e631c854b0d9edd5d93e91d5b8c"

	got, err := dispatcher.DefineAndLaunch(dp)
	require.Nil(t, err)
	require.Equal(t, want, got.ID)
}
