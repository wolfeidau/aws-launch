package ecs

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/wolfeidau/aws-launch/awsmocks"
	"github.com/wolfeidau/aws-launch/mocks"
	"github.com/wolfeidau/aws-launch/pkg/cwlogs"
	"github.com/wolfeidau/aws-launch/pkg/launcher"
)

func Test_ShortenARN(t *testing.T) {
	v := shortenTaskArn(aws.String("arn:aws:ecs:ap-southeast-2:123456789012:task/wolfeidau-ecs-dev-Cluster-1234567890123/abcefg1234567890abcefg1234567890"))
	require.Equal(t, "abcefg1234567890abcefg1234567890", v)
}

func TestLauncher_LaunchTask(t *testing.T) {

	cwlogsSvcMock := &awsmocks.CloudWatchLogsAPI{}
	ecsSvcMock := &awsmocks.ECSAPI{}

	ecsSvcMock.On("RunTask", mock.AnythingOfType("*ecs.RunTaskInput")).Return(&ecs.RunTaskOutput{
		Tasks: []*ecs.Task{
			{
				TaskArn: aws.String("arn:aws:ecs:ap-southeast-2:123456789012:task/wolfeidau-ecs-dev-Cluster-1234567890123/dece5e631c854b0d9edd5d93e91d5b8c"),
			},
		},
	}, nil)

	rt := &LaunchTaskParams{
		ClusterName:    "abc123",
		ContainerName:  "test-command",
		ServiceName:    "test-command",
		TaskDefinition: "test-command:12",
	}

	want := &LaunchTaskResult{
		ID:         "arn:aws:ecs:ap-southeast-2:123456789012:task/wolfeidau-ecs-dev-Cluster-1234567890123/dece5e631c854b0d9edd5d93e91d5b8c",
		TaskStatus: launcher.TaskRunning,
		TaskArn:    "arn:aws:ecs:ap-southeast-2:123456789012:task/wolfeidau-ecs-dev-Cluster-1234567890123/dece5e631c854b0d9edd5d93e91d5b8c",
		TaskID:     "dece5e631c854b0d9edd5d93e91d5b8c",
	}
	cbl := &Launcher{
		ecsSvc:    ecsSvcMock,
		cwlogsSvc: cwlogsSvcMock,
	}
	got, err := cbl.LaunchTask(rt)
	require.Nil(t, err)
	require.Equal(t, want, got)

}

func TestLauncher_DefineTask_With_Update(t *testing.T) {

	cwlogsSvcMock := &awsmocks.CloudWatchLogsAPI{}
	ecsSvcMock := &awsmocks.ECSAPI{}

	cwlogsSvcMock.On("CreateLogGroup", mock.AnythingOfType("*cloudwatchlogs.CreateLogGroupInput")).Return(&cloudwatchlogs.CreateLogGroupOutput{}, nil)
	ecsSvcMock.On("RegisterTaskDefinition", mock.AnythingOfType("*ecs.RegisterTaskDefinitionInput")).Return(&ecs.RegisterTaskDefinitionOutput{
		TaskDefinition: &ecs.TaskDefinition{
			Family:   aws.String("test-command"),
			Revision: aws.Int64(123),
		},
	}, nil)

	dp := &DefineTaskParams{
		ContainerName:    "test-command",
		DefinitionName:   "test-command",
		ExecutionRoleARN: "arn:aws:iam::123456789012:role/ecsTaskExecutionRole",
		Image:            "wolfeidau/test-command:latest",
		Region:           "ap-southeast-2",
	}
	want := &DefineTaskResult{
		ID:                     "test-command:123",
		CloudwatchLogGroupName: "/aws/fargate/test-command",
		CloudwatchStreamPrefix: "ecs",
	}

	cbl := &Launcher{
		ecsSvc:    ecsSvcMock,
		cwlogsSvc: cwlogsSvcMock,
	}

	got, err := cbl.DefineTask(dp)
	require.Nil(t, err)
	require.Equal(t, want, got)
}

func TestLauncher_GetTaskStatus(t *testing.T) {

	ecsSvcMock := &awsmocks.ECSAPI{}

	ecsSvcMock.On("DescribeTasks", mock.AnythingOfType("*ecs.DescribeTasksInput")).Return(&ecs.DescribeTasksOutput{
		Tasks: []*ecs.Task{
			{
				LastStatus: aws.String(ecs.DesiredStatusStopped),
				StopCode:   aws.String("EssentialContainerExited"),
				TaskArn:    aws.String("arn:aws:ecs:ap-southeast-2:123456789012:task/wolfeidau-ecs-dev-Cluster-1234567890123/dece5e631c854b0d9edd5d93e91d5b8c"),
			},
		},
	}, nil)

	gt := &GetTaskStatusParams{
		ClusterName: "testing-1",
	}
	want := &GetTaskStatusResult{
		TaskArn:    "arn:aws:ecs:ap-southeast-2:123456789012:task/wolfeidau-ecs-dev-Cluster-1234567890123/dece5e631c854b0d9edd5d93e91d5b8c",
		TaskID:     "dece5e631c854b0d9edd5d93e91d5b8c",
		ID:         "arn:aws:ecs:ap-southeast-2:123456789012:task/wolfeidau-ecs-dev-Cluster-1234567890123/dece5e631c854b0d9edd5d93e91d5b8c",
		TaskStatus: "SUCCEEDED",
	}

	cbl := &Launcher{
		ecsSvc: ecsSvcMock,
	}

	got, err := cbl.GetTaskStatus(gt)
	require.Nil(t, err)
	require.Equal(t, want, got)
}

func TestLauncher_CleanupTask(t *testing.T) {

	ecsSvcMock := &awsmocks.ECSAPI{}

	ecsSvcMock.On("DeregisterTaskDefinition", mock.AnythingOfType("*ecs.DeregisterTaskDefinitionInput")).Return(&ecs.DeregisterTaskDefinitionOutput{}, nil)

	ct := &CleanupTaskParams{
		TaskDefinition: "test-command:12",
	}
	want := &CleanupTaskResult{}

	cbl := &Launcher{
		ecsSvc: ecsSvcMock,
	}

	got, err := cbl.CleanupTask(ct)
	require.Nil(t, err)
	require.Equal(t, want, got)
}

func TestLauncher_GetTaskLogs(t *testing.T) {

	cwlogsReader := &mocks.LogsReader{}

	cwlogsReader.On("ReadLogs", mock.AnythingOfType("*cwlogs.ReadLogsParams")).Return(&cwlogs.ReadLogsResult{
		LogLines:  []*cwlogs.LogLine{{Message: "whaterer"}},
		NextToken: aws.String("f/123456789"),
	}, nil)

	gt := &GetTaskLogsParams{}

	want := &GetTaskLogsResult{
		LogLines:  []*cwlogs.LogLine{{Message: "whaterer"}},
		NextToken: aws.String("f/123456789"),
	}

	cbl := &Launcher{
		cwlogsReader: cwlogsReader,
	}

	got, err := cbl.GetTaskLogs(gt)
	require.Nil(t, err)
	require.Equal(t, want, got)
}
