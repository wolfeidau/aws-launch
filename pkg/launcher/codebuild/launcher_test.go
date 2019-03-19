package codebuild

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/wolfeidau/aws-launch/awsmocks"
	"github.com/wolfeidau/aws-launch/mocks"
	"github.com/wolfeidau/aws-launch/pkg/cwlogs"
	"github.com/wolfeidau/aws-launch/pkg/launcher"
)

const codebuildArn = "arn:aws:codebuild:ap-southeast-2:123456789012:build/BuildkiteProject-dev-1:b17dddde-97c6-4592-b7be-216524f8422b"

func TestLauncher_LaunchTask(t *testing.T) {

	cwlogsSvcMock := &awsmocks.CloudWatchLogsAPI{}
	codeBuildSvcMock := &awsmocks.CodeBuildAPI{}

	codeBuildSvcMock.On("StartBuild", mock.AnythingOfType("*codebuild.StartBuildInput")).Return(&codebuild.StartBuildOutput{
		Build: &codebuild.Build{
			Id:          aws.String("abc123"),
			BuildStatus: aws.String(codebuild.StatusTypeInProgress),
			Arn:         aws.String(codebuildArn),
		},
	}, nil)

	rt := &LaunchTaskParams{
		ProjectName: "testing-1",
		Environment: map[string]string{
			"TestEnv": "test",
		},
	}

	want := &LaunchTaskResult{
		ID:          "abc123",
		TaskStatus:  launcher.TaskRunning,
		BuildArn:    codebuildArn,
		BuildStatus: codebuild.StatusTypeInProgress,
	}
	cbl := &Launcher{
		codeBuildSvc: codeBuildSvcMock,
		cwlogsSvc:    cwlogsSvcMock,
	}
	got, err := cbl.LaunchTask(rt)
	require.Nil(t, err)
	require.Equal(t, want, got)

}

func TestLauncher_DefineTask_With_Update(t *testing.T) {

	cwlogsSvcMock := &awsmocks.CloudWatchLogsAPI{}
	codeBuildSvcMock := &awsmocks.CodeBuildAPI{}

	cwlogsSvcMock.On("CreateLogGroup", mock.AnythingOfType("*cloudwatchlogs.CreateLogGroupInput")).Return(&cloudwatchlogs.CreateLogGroupOutput{}, nil)
	codeBuildSvcMock.On("UpdateProject", mock.AnythingOfType("*codebuild.UpdateProjectInput")).Return(&codebuild.UpdateProjectOutput{
		Project: &codebuild.Project{
			Arn: aws.String("abc123/codebuild/whatever"),
		},
	}, nil)

	dp := &DefineTaskParams{
		ProjectName: "testing-1",
		ComputeType: "BUILD_GENERAL1_SMALL",
		ServiceRole: "abc123Role",
		Tags: map[string]string{
			"TestTag": "test",
		},
		Environment: map[string]string{
			"TestEnv": "test",
		},
	}
	want := &DefineTaskResult{
		ID:                     "abc123/codebuild/whatever",
		CloudwatchLogGroupName: "/aws/codebuild/testing-1",
		CloudwatchStreamPrefix: "codebuild",
	}

	cbl := &Launcher{
		codeBuildSvc: codeBuildSvcMock,
		cwlogsSvc:    cwlogsSvcMock,
	}

	got, err := cbl.DefineTask(dp)
	require.Nil(t, err)
	require.Equal(t, want, got)
}

func TestLauncher_GetTaskStatus(t *testing.T) {

	codeBuildSvcMock := &awsmocks.CodeBuildAPI{}

	codeBuildSvcMock.On("BatchGetBuilds", mock.AnythingOfType("*codebuild.BatchGetBuildsInput")).Return(&codebuild.BatchGetBuildsOutput{
		Builds: []*codebuild.Build{
			{
				BuildStatus: aws.String(codebuild.StatusTypeSucceeded),
				Arn:         aws.String(codebuildArn),
			},
		},
	}, nil)

	gt := &GetTaskStatusParams{
		ID: "testing-1",
	}
	want := &GetTaskStatusResult{
		BuildArn:    codebuildArn,
		BuildStatus: "SUCCEEDED",
		ID:          codebuildArn,
		TaskStatus:  launcher.TaskSucceeded,
	}

	cbl := &Launcher{
		codeBuildSvc: codeBuildSvcMock,
	}

	got, err := cbl.GetTaskStatus(gt)
	require.Nil(t, err)
	require.Equal(t, want, got)
}

func TestLauncher_CleanUpTask(t *testing.T) {

	codeBuildSvcMock := &awsmocks.CodeBuildAPI{}

	codeBuildSvcMock.On("DeleteProject", mock.AnythingOfType("*codebuild.DeleteProjectInput")).Return(&codebuild.DeleteProjectOutput{}, nil)

	ct := &CleanupTaskParams{
		ProjectName: "testing-1",
	}
	want := &CleanupTaskResult{}

	cbl := &Launcher{
		codeBuildSvc: codeBuildSvcMock,
	}

	got, err := cbl.CleanupTask(ct)
	require.Nil(t, err)
	require.Equal(t, want, got)
}

func TestLauncher_GetTaskLogs(t *testing.T) {

	cwlogsReader := &mocks.LogsReader{}

	cwlogsReader.On("ReadLogs", mock.AnythingOfType("*cwlogs.ReadLogsParams")).Return(&cwlogs.ReadLogsResult{
		LogLines:  []*cwlogs.LogLine{{Message: "whatever"}},
		NextToken: aws.String("f/123456789"),
	}, nil)

	gt := &GetTaskLogsParams{}

	want := &GetTaskLogsResult{
		LogLines:  []*cwlogs.LogLine{{Message: "whatever"}},
		NextToken: aws.String("f/123456789"),
	}

	cbl := &Launcher{
		cwlogsReader: cwlogsReader,
	}

	got, err := cbl.GetTaskLogs(gt)
	require.Nil(t, err)
	require.Equal(t, want, got)
}
