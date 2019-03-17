package codebuild

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/wolfeidau/aws-launch/mocks"
	"github.com/wolfeidau/aws-launch/pkg/cwlogs"
	"github.com/wolfeidau/aws-launch/pkg/launcher"
)

const codebuildArn = "arn:aws:codebuild:ap-southeast-2:123456789012:build/BuildkiteProject-dev-1:b17dddde-97c6-4592-b7be-216524f8422b"

func TestCodeBuildLauncher_DefineAndLaunchTask(t *testing.T) {
	cwlogsSvcMock := &mocks.CloudWatchLogsAPI{}
	codeBuildSvcMock := &mocks.CodeBuildAPI{}

	codeBuildSvcMock.On("StartBuild", mock.AnythingOfType("*codebuild.StartBuildInput")).Return(&codebuild.StartBuildOutput{
		Build: &codebuild.Build{
			Id:          aws.String("abc123"),
			BuildStatus: aws.String(codebuild.StatusTypeInProgress),
			Arn:         aws.String(codebuildArn),
		},
	}, nil)

	cwlogsSvcMock.On("CreateLogGroup", mock.AnythingOfType("*cloudwatchlogs.CreateLogGroupInput")).Return(nil, awserr.New("ResourceAlreadyExistsException", "", nil))
	codeBuildSvcMock.On("UpdateProject", mock.AnythingOfType("*codebuild.UpdateProjectInput")).Return(nil, awserr.New("ResourceNotFoundException", "", nil))
	codeBuildSvcMock.On("CreateProject", mock.AnythingOfType("*codebuild.CreateProjectInput")).Return(&codebuild.CreateProjectOutput{
		Project: &codebuild.Project{
			Arn: aws.String(codebuildArn),
		},
	}, nil)

	dl := &launcher.DefineAndLaunchParams{
		Codebuild: &launcher.CodebuildDefineAndLaunchParams{
			ProjectName: "testing-1",
			ComputeType: "BUILD_GENERAL1_SMALL",
			ServiceRole: "abc123Role",
		},
		Environment: map[string]string{
			"TestEnv": "test",
		},
		Tags: map[string]string{
			"TestTag": "test",
		},
	}

	want := &launcher.DefineAndLaunchResult{
		BaseTaskResult: &launcher.BaseTaskResult{
			ID: "abc123",
			CodeBuild: &launcher.LaunchTaskCodebuildResult{
				BuildArn:    codebuildArn,
				BuildStatus: codebuild.StatusTypeInProgress,
			},
			TaskStatus: launcher.TaskRunning,
		},
		DefinitionID:           codebuildArn,
		CloudwatchLogGroupName: "/aws/codebuild/testing-1",
		CloudwatchStreamPrefix: "codebuild",
	}
	cbl := &CodeBuildLauncher{
		codeBuildSvc: codeBuildSvcMock,
		cwlogsSvc:    cwlogsSvcMock,
	}
	got, err := cbl.DefineAndLaunch(dl)
	require.Nil(t, err)
	require.Equal(t, want, got)
}

func TestCodeBuildLauncher_LaunchTask(t *testing.T) {

	cwlogsSvcMock := &mocks.CloudWatchLogsAPI{}
	codeBuildSvcMock := &mocks.CodeBuildAPI{}

	codeBuildSvcMock.On("StartBuild", mock.AnythingOfType("*codebuild.StartBuildInput")).Return(&codebuild.StartBuildOutput{
		Build: &codebuild.Build{
			Id:          aws.String("abc123"),
			BuildStatus: aws.String(codebuild.StatusTypeInProgress),
			Arn:         aws.String(codebuildArn),
		},
	}, nil)

	rt := &launcher.LaunchTaskParams{
		Codebuild: &launcher.CodebuildTaskParams{
			ProjectName: "testing-1",
		},
		Environment: map[string]string{
			"TestEnv": "test",
		},
	}

	want := &launcher.LaunchTaskResult{
		BaseTaskResult: &launcher.BaseTaskResult{
			ID:         "abc123",
			TaskStatus: launcher.TaskRunning,
			CodeBuild: &launcher.LaunchTaskCodebuildResult{
				BuildArn:    codebuildArn,
				BuildStatus: codebuild.StatusTypeInProgress,
			},
		},
	}
	cbl := &CodeBuildLauncher{
		codeBuildSvc: codeBuildSvcMock,
		cwlogsSvc:    cwlogsSvcMock,
	}
	got, err := cbl.LaunchTask(rt)
	require.Nil(t, err)
	require.Equal(t, want, got)

}

func TestCodeBuildLauncher_DefineTask_With_Update(t *testing.T) {

	cwlogsSvcMock := &mocks.CloudWatchLogsAPI{}
	codeBuildSvcMock := &mocks.CodeBuildAPI{}

	cwlogsSvcMock.On("CreateLogGroup", mock.AnythingOfType("*cloudwatchlogs.CreateLogGroupInput")).Return(&cloudwatchlogs.CreateLogGroupOutput{}, nil)
	codeBuildSvcMock.On("UpdateProject", mock.AnythingOfType("*codebuild.UpdateProjectInput")).Return(&codebuild.UpdateProjectOutput{
		Project: &codebuild.Project{
			Arn: aws.String("abc123/codebuild/whatever"),
		},
	}, nil)

	dp := &launcher.DefineTaskParams{
		Codebuild: &launcher.CodebuildDefineTaskParams{
			ProjectName: "testing-1",
			ComputeType: "BUILD_GENERAL1_SMALL",
			ServiceRole: "abc123Role",
		},
		Tags: map[string]string{
			"TestTag": "test",
		},
		Environment: map[string]string{
			"TestEnv": "test",
		},
	}
	want := &launcher.DefineTaskResult{
		ID:                     "abc123/codebuild/whatever",
		CloudwatchLogGroupName: "/aws/codebuild/testing-1",
		CloudwatchStreamPrefix: "codebuild",
	}

	cbl := &CodeBuildLauncher{
		codeBuildSvc: codeBuildSvcMock,
		cwlogsSvc:    cwlogsSvcMock,
	}

	got, err := cbl.DefineTask(dp)
	require.Nil(t, err)
	require.Equal(t, want, got)
}

func TestCodeBuildLauncher_GetTaskStatus(t *testing.T) {

	codeBuildSvcMock := &mocks.CodeBuildAPI{}

	codeBuildSvcMock.On("BatchGetBuilds", mock.AnythingOfType("*codebuild.BatchGetBuildsInput")).Return(&codebuild.BatchGetBuildsOutput{
		Builds: []*codebuild.Build{
			{
				BuildStatus: aws.String(codebuild.StatusTypeSucceeded),
				Arn:         aws.String(codebuildArn),
			},
		},
	}, nil)

	gt := &launcher.GetTaskStatusParams{
		Codebuild: &launcher.CodebuildTaskParams{
			ProjectName: "testing-1",
		},
	}
	want := &launcher.GetTaskStatusResult{
		BaseTaskResult: &launcher.BaseTaskResult{
			CodeBuild: &launcher.LaunchTaskCodebuildResult{
				BuildArn:    codebuildArn,
				BuildStatus: "SUCCEEDED",
			},
			ID:         codebuildArn,
			TaskStatus: launcher.TaskSucceeded,
		},
	}

	cbl := &CodeBuildLauncher{
		codeBuildSvc: codeBuildSvcMock,
	}

	got, err := cbl.GetTaskStatus(gt)
	require.Nil(t, err)
	require.Equal(t, want, got)
}

func TestCodeBuildLauncher_CleanUpTask(t *testing.T) {

	codeBuildSvcMock := &mocks.CodeBuildAPI{}

	codeBuildSvcMock.On("DeleteProject", mock.AnythingOfType("*codebuild.DeleteProjectInput")).Return(&codebuild.DeleteProjectOutput{}, nil)

	ct := &launcher.CleanupTaskParams{
		Codebuild: &launcher.CodebuildCleanupTaskParams{
			ProjectName: "testing-1",
		},
	}
	want := &launcher.CleanupTaskResult{}

	cbl := &CodeBuildLauncher{
		codeBuildSvc: codeBuildSvcMock,
	}

	got, err := cbl.CleanupTask(ct)
	require.Nil(t, err)
	require.Equal(t, want, got)
}

func TestCodeBuildLauncher_GetTaskLogs(t *testing.T) {

	cwlogsReader := &mocks.LogsReader{}

	cwlogsReader.On("ReadLogs", mock.AnythingOfType("*cwlogs.ReadLogsParams")).Return(&cwlogs.ReadLogsResult{
		LogLines:  []*cwlogs.LogLine{{Message: "whatever"}},
		NextToken: aws.String("f/123456789"),
	}, nil)

	gt := &launcher.GetTaskLogsParams{
		Codebuild: &launcher.CodebuildTaskLogsParams{},
	}

	want := &launcher.GetTaskLogsResult{
		LogLines:  []*cwlogs.LogLine{{Message: "whatever"}},
		NextToken: aws.String("f/123456789"),
	}

	cbl := &CodeBuildLauncher{
		cwlogsReader: cwlogsReader,
	}

	got, err := cbl.GetTaskLogs(gt)
	require.Nil(t, err)
	require.Equal(t, want, got)
}
