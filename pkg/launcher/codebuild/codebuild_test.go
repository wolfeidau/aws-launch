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
)

func TestCodeBuildLauncher_DefineAndLaunchTask(t *testing.T) {
	cwlogsSvcMock := &mocks.CloudWatchLogsAPI{}
	codeBuildSvcMock := &mocks.CodeBuildAPI{}

	codeBuildSvcMock.On("StartBuild", mock.AnythingOfType("*codebuild.StartBuildInput")).Return(&codebuild.StartBuildOutput{
		Build: &codebuild.Build{
			Id: aws.String("abc123"),
		},
	}, nil)

	cwlogsSvcMock.On("CreateLogGroup", mock.AnythingOfType("*cloudwatchlogs.CreateLogGroupInput")).Return(nil, awserr.New("ResourceAlreadyExistsException", "", nil))
	codeBuildSvcMock.On("UpdateProject", mock.AnythingOfType("*codebuild.UpdateProjectInput")).Return(nil, awserr.New("ResourceNotFoundException", "", nil))
	codeBuildSvcMock.On("CreateProject", mock.AnythingOfType("*codebuild.CreateProjectInput")).Return(&codebuild.CreateProjectOutput{
		Project: &codebuild.Project{
			Arn: aws.String("abc123/codebuild/whatever"),
		},
	}, nil)

	dl := &DefineAndLaunchParams{
		Codebuild: &CodebuildDefineAndLaunchParams{
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

	want := &DefineAndLaunchResult{
		BaseTaskResult: &BaseTaskResult{
			ID: "abc123",
		},
		DefinitionID:           "abc123/codebuild/whatever",
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
			Id: aws.String("abc123"),
		},
	}, nil)

	rt := &LaunchTaskParams{
		Codebuild: &CodebuildTaskParams{
			ProjectName: "testing-1",
		},
		Environment: map[string]string{
			"TestEnv": "test",
		},
	}

	want := &LaunchTaskResult{
		BaseTaskResult: &BaseTaskResult{
			ID: "abc123",
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

	dp := &DefineTaskParams{
		Codebuild: &CodebuildDefineTaskParams{
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
	want := &DefineTaskResult{
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
				Arn:         aws.String("abc123"),
			},
		},
	}, nil)

	gt := &GetTaskStatusParams{
		Codebuild: &CodebuildTaskParams{
			ProjectName: "testing-1",
		},
	}
	want := &GetTaskStatusResult{
		BaseTaskResult: &BaseTaskResult{
			CodeBuild: &LaunchTaskCodebuildResult{
				BuildArn:    "abc123",
				BuildStatus: "SUCCEEDED",
			},
			ID:         "abc123",
			TaskStatus: "RUNNING",
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

	ct := &CleanupTaskParams{
		Codebuild: &CodebuildCleanupTaskParams{
			ProjectName: "testing-1",
		},
	}
	want := &CleanupTaskResult{}

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

	gt := &GetTaskLogsParams{
		Codebuild: &CodebuildTaskLogsParams{},
	}

	want := &GetTaskLogsResult{
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
