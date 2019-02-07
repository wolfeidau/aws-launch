package launcher

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/wolfeidau/fargate-run-job/mocks"
	"github.com/wolfeidau/fargate-run-job/pkg/cwlogs"
)

func TestCodeBuildLauncher_LaunchTask(t *testing.T) {

	cwlogsSvcMock := &mocks.CloudWatchLogsAPI{}
	codeBuildSvcMock := &mocks.CodeBuildAPI{}

	codeBuildSvcMock.On("StartBuild", mock.AnythingOfType("*codebuild.StartBuildInput")).Return(&codebuild.StartBuildOutput{
		Build: &codebuild.Build{
			Id: aws.String("abc123"),
		},
	}, nil)

	cwlogsSvc := cwlogsSvcMock
	codeBuildSvc := codeBuildSvcMock

	rt := &LaunchTaskParams{
		Codebuild: &CodebuildTaskParams{
			ProjectName: "testing-1",
		},
	}

	want := &LaunchTaskResult{
		BaseTaskResult: &BaseTaskResult{
			ID: "abc123",
		},
	}
	cbl := &CodeBuildLauncher{
		codeBuildSvc: codeBuildSvc,
		cwlogsSvc:    cwlogsSvc,
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

	cwlogsSvc := cwlogsSvcMock
	codeBuildSvc := codeBuildSvcMock

	dp := &DefineTaskParams{
		Codebuild: &CodebuildDefineTaskParams{
			ProjectName: "testing-1",
			ComputeType: "BUILD_GENERAL1_SMALL",
			ServiceRole: "abc123Role",
		},
	}
	want := &DefineTaskResult{
		ID: "abc123/codebuild/whatever",
	}

	cbl := &CodeBuildLauncher{
		codeBuildSvc: codeBuildSvc,
		cwlogsSvc:    cwlogsSvc,
	}

	got, err := cbl.DefineTask(dp)
	require.Nil(t, err)
	require.Equal(t, want, got)
}

func TestCodeBuildLauncher_GetTaskLogs(t *testing.T) {

	cwlogsReader := &mocks.LogsReader{}

	cwlogsReader.On("ReadLogs", mock.AnythingOfType("*cwlogs.ReadLogsParams")).Return(&cwlogs.ReadLogsResult{
		LogLines:  []*cwlogs.LogLine{&cwlogs.LogLine{Message: "whaterer"}},
		NextToken: aws.String("f/123456789"),
	}, nil)

	gt := &GetTaskLogsParams{
		Codebuild: &CodebuildTaskLogsParams{},
	}

	want := &GetTaskLogsResult{
		LogLines:  []*cwlogs.LogLine{&cwlogs.LogLine{Message: "whaterer"}},
		NextToken: aws.String("f/123456789"),
	}

	cbl := &CodeBuildLauncher{
		cwlogsReader: cwlogsReader,
	}

	got, err := cbl.GetTaskLogs(gt)
	require.Nil(t, err)
	require.Equal(t, want, got)
}
