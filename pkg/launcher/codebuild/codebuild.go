package codebuild

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/codebuild/codebuildiface"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wolfeidau/aws-launch/pkg/cwlogs"
	"github.com/wolfeidau/aws-launch/pkg/launcher"
)

const (
	// CodebuildStreamPrefix the prefix used in the cloudwatch log stream name
	CodebuildStreamPrefix = "codebuild"

	// CodebuildLogGroupFormat the name format for cloudwatch log group names
	CodebuildLogGroupFormat = "/aws/codebuild/%s"
)

// CodeBuildLauncher used to launch containers in CodeBuild
type CodeBuildLauncher struct {
	codeBuildSvc codebuildiface.CodeBuildAPI
	cwlogsSvc    cloudwatchlogsiface.CloudWatchLogsAPI
	cwlogsReader cwlogs.LogsReader
}

// NewCodeBuildLauncher create a new launcher
func NewCodeBuildLauncher(cfgs ...*aws.Config) *CodeBuildLauncher {
	sess := session.Must(session.NewSession(cfgs...))
	return &CodeBuildLauncher{
		codeBuildSvc: codebuild.New(sess),
		cwlogsSvc:    cloudwatchlogs.New(sess),
		cwlogsReader: cwlogs.NewCloudwatchLogsReader(cfgs...),
	}
}

// DefineTask create or update a codebuild job for this definition and return the ARN of this job
func (cbl *CodeBuildLauncher) DefineTask(dp *launcher.DefineTaskParams) (*launcher.DefineTaskResult, error) {

	logGroupName := fmt.Sprintf(CodebuildLogGroupFormat, dp.Codebuild.ProjectName)

	_, err := cbl.cwlogsSvc.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName),
		Tags: map[string]*string{
			"createdBy": aws.String("fargate-run-job"),
		},
	})
	if err, ok := err.(awserr.Error); ok {
		if err.Code() != "ResourceAlreadyExistsException" {
			return nil, errors.Wrap(err, "create log group failed.")
		}

		logrus.WithField("name", logGroupName).Info("cloudwatch log group exists")
	}

	// just update the project to see if it already exists
	// NOTE: currently codebuild list projects call has no filter
	projectArn, updated, err := cbl.tryUpdateProject(dp, logGroupName)
	if err != nil {
		return nil, err
	}
	if updated {
		logrus.WithField("projectArn", projectArn).Info("updated codebuild project")

		return &launcher.DefineTaskResult{
			ID:                     projectArn,
			CloudwatchLogGroupName: logGroupName,
			CloudwatchStreamPrefix: "codebuild",
		}, nil
	}

	createRes, err := cbl.codeBuildSvc.CreateProject(&codebuild.CreateProjectInput{
		Name: aws.String(dp.Codebuild.ProjectName),
		Environment: &codebuild.ProjectEnvironment{
			ComputeType:          aws.String(dp.Codebuild.ComputeType),
			Image:                aws.String(dp.Image),
			Type:                 aws.String(codebuild.EnvironmentTypeLinuxContainer),
			PrivilegedMode:       dp.Codebuild.PrivilegedMode,
			EnvironmentVariables: convertMapToEnvironmentVariable(dp.Environment),
		},
		Artifacts: &codebuild.ProjectArtifacts{
			Type: aws.String(codebuild.ArtifactsTypeNoArtifacts),
		},
		Source: &codebuild.ProjectSource{
			Type:      aws.String(codebuild.SourceTypeNoSource),
			Buildspec: aws.String(dp.Codebuild.Buildspec),
		},
		ServiceRole: aws.String(dp.Codebuild.ServiceRole),
		LogsConfig: &codebuild.LogsConfig{
			CloudWatchLogs: &codebuild.CloudWatchLogsConfig{
				GroupName:  aws.String(logGroupName),
				StreamName: aws.String(CodebuildStreamPrefix),
				Status:     aws.String(codebuild.LogsConfigStatusTypeEnabled),
			},
		},
		Tags: convertMapToCodebuildTags(dp.Tags),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to register project.")
	}

	projectArn = aws.StringValue(createRes.Project.Arn)

	logrus.WithField("projectArn", projectArn).Info("created codebuild project")

	return &launcher.DefineTaskResult{
		ID:                     projectArn,
		CloudwatchLogGroupName: logGroupName,
		CloudwatchStreamPrefix: "codebuild",
	}, nil
}

// LaunchTask run a container task and monitor it till completion
func (cbl *CodeBuildLauncher) LaunchTask(rt *launcher.LaunchTaskParams) (*launcher.LaunchTaskResult, error) {

	res, err := cbl.codeBuildSvc.StartBuild(&codebuild.StartBuildInput{
		ProjectName:                  aws.String(rt.Codebuild.ProjectName),
		EnvironmentVariablesOverride: convertMapToEnvironmentVariable(rt.Environment),
		ImageOverride:                rt.Codebuild.Image,
		ComputeTypeOverride:          rt.Codebuild.ComputeType,
		PrivilegedModeOverride:       rt.Codebuild.PrivilegedMode,
		ServiceRoleOverride:          rt.Codebuild.ServiceRole,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to start build.")
	}

	taskRes := &launcher.LaunchTaskResult{
		ID:         aws.StringValue(res.Build.Id),
		TaskStatus: convertTaskStatus(aws.StringValue(res.Build.BuildStatus)),
		CodeBuild: &launcher.LaunchTaskCodebuildResult{
			BuildArn:    aws.StringValue(res.Build.Arn),
			BuildStatus: aws.StringValue(res.Build.BuildStatus),
		},
	}

	return taskRes, nil
}

// WaitForTask wait for task to complete
func (cbl *CodeBuildLauncher) WaitForTask(wft *launcher.WaitForTaskParams) (*launcher.WaitForTaskResult, error) {

	params := &codebuild.BatchGetBuildsInput{
		Ids: []*string{aws.String(wft.ID)},
	}

	err := cbl.waitUntilTasksStoppedWithContext(context.Background(), params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start build.")
	}

	return &launcher.WaitForTaskResult{ID: wft.ID}, nil
}

// GetTaskStatus get task status
func (cbl *CodeBuildLauncher) GetTaskStatus(gts *launcher.GetTaskStatusParams) (*launcher.GetTaskStatusResult, error) {

	params := &codebuild.BatchGetBuildsInput{
		Ids: []*string{aws.String(gts.ID)},
	}
	getBuildRes, err := cbl.codeBuildSvc.BatchGetBuilds(params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start build.")
	}

	build := getBuildRes.Builds[0]

	logrus.WithFields(logrus.Fields{
		"BuildComplete": aws.BoolValue(build.BuildComplete),
		"BuildStatus":   aws.StringValue(build.BuildStatus),
		"StartTime":     aws.TimeValue(build.StartTime),
		"StopTime":      aws.TimeValue(build.EndTime),
	}).Info("Describe completed Task")

	taskRes := &launcher.GetTaskStatusResult{
		ID:         aws.StringValue(build.Arn),
		StartTime:  build.StartTime,
		EndTime:    build.EndTime,
		TaskStatus: convertTaskStatus(aws.StringValue(build.BuildStatus)),
		CodeBuild: &launcher.LaunchTaskCodebuildResult{
			BuildArn:    aws.StringValue(build.Arn),
			BuildStatus: aws.StringValue(build.BuildStatus),
		},
	}

	if aws.BoolValue(build.BuildComplete) {
		if aws.StringValue(build.BuildStatus) == "SUCCEEDED" {
			taskRes.TaskStatus = launcher.TaskSucceeded
		} else {
			taskRes.TaskStatus = launcher.TaskFailed
		}
	}

	return taskRes, nil

}

// CleanupTask clean up codebuild project
func (cbl *CodeBuildLauncher) CleanupTask(ctp *launcher.CleanupTaskParams) (*launcher.CleanupTaskResult, error) {
	_, err := cbl.codeBuildSvc.DeleteProject(&codebuild.DeleteProjectInput{
		Name: aws.String(ctp.Codebuild.ProjectName),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to delete project.")
	}

	return &launcher.CleanupTaskResult{}, nil
}

// GetTaskLogs get task logs
func (cbl *CodeBuildLauncher) GetTaskLogs(gtlp *launcher.GetTaskLogsParams) (*launcher.GetTaskLogsResult, error) {

	logGroupName := fmt.Sprintf(CodebuildLogGroupFormat, gtlp.Codebuild.ProjectName)
	streamName := fmt.Sprintf("%s/%s", CodebuildStreamPrefix, gtlp.Codebuild.TaskID)

	res, err := cbl.cwlogsReader.ReadLogs(&cwlogs.ReadLogsParams{
		GroupName:  logGroupName,
		StreamName: streamName,
		NextToken:  gtlp.NextToken,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve logs for task.")
	}

	return &launcher.GetTaskLogsResult{
		LogLines:  res.LogLines,
		NextToken: res.NextToken,
	}, nil
}

func (cbl *CodeBuildLauncher) tryUpdateProject(dp *launcher.DefineTaskParams, logGroupName string) (string, bool, error) {
	updateRes, err := cbl.codeBuildSvc.UpdateProject(&codebuild.UpdateProjectInput{
		Name: aws.String(dp.Codebuild.ProjectName),
		Environment: &codebuild.ProjectEnvironment{
			ComputeType:          aws.String(dp.Codebuild.ComputeType),
			Image:                aws.String(dp.Image),
			Type:                 aws.String(codebuild.EnvironmentTypeLinuxContainer),
			PrivilegedMode:       dp.Codebuild.PrivilegedMode,
			EnvironmentVariables: convertMapToEnvironmentVariable(dp.Environment),
		},
		Artifacts: &codebuild.ProjectArtifacts{
			Type: aws.String(codebuild.ArtifactsTypeNoArtifacts),
		},
		Source: &codebuild.ProjectSource{
			Type:      aws.String(codebuild.SourceTypeNoSource),
			Buildspec: aws.String(dp.Codebuild.Buildspec),
		},
		ServiceRole: aws.String(dp.Codebuild.ServiceRole),
		LogsConfig: &codebuild.LogsConfig{
			CloudWatchLogs: &codebuild.CloudWatchLogsConfig{
				GroupName:  aws.String(logGroupName),
				StreamName: aws.String("codebuild"),
				Status:     aws.String(codebuild.LogsConfigStatusTypeEnabled),
			},
		},
		Tags: convertMapToCodebuildTags(dp.Tags),
	})
	if err, ok := err.(awserr.Error); ok {
		if err.Code() == "ResourceNotFoundException" {
			return "", false, nil // skip this error as the job will be subsequently created
		}
		return "", false, errors.Wrap(err, "update codebuild project failed.")
	}

	return aws.StringValue(updateRes.Project.Arn), true, nil
}

func (cbl *CodeBuildLauncher) waitUntilTasksStoppedWithContext(ctx aws.Context, input *codebuild.BatchGetBuildsInput, opts ...request.WaiterOption) error {
	w := request.Waiter{
		Name:        "WaitUntilBuildsStopped",
		MaxAttempts: 100,
		Delay:       request.ConstantWaiterDelay(6 * time.Second),
		Acceptors: []request.WaiterAcceptor{
			{
				State:   request.SuccessWaiterState,
				Matcher: request.PathAllWaiterMatch, Argument: "builds[].buildComplete",
				Expected: true,
			},
		},
		NewRequest: func(opts []request.Option) (*request.Request, error) {
			var inCpy *codebuild.BatchGetBuildsInput
			if input != nil {
				tmp := *input
				inCpy = &tmp
			}
			req, _ := cbl.codeBuildSvc.BatchGetBuildsRequest(inCpy)
			req.SetContext(ctx)
			req.ApplyOptions(opts...)
			return req, nil
		},
	}
	w.ApplyOptions(opts...)

	return w.WaitWithContext(ctx)
}

func convertMapToEnvironmentVariable(env map[string]string) []*codebuild.EnvironmentVariable {

	codebuildEnv := []*codebuild.EnvironmentVariable{}

	// empty map is valid
	if env == nil {
		return nil
	}

	for k, v := range env {
		codebuildEnv = append(codebuildEnv, &codebuild.EnvironmentVariable{Name: aws.String(k), Value: aws.String(v)})
	}

	return codebuildEnv
}

func convertMapToCodebuildTags(tags map[string]string) []*codebuild.Tag {

	codebuildTags := []*codebuild.Tag{}

	// empty map is valid
	if tags == nil {
		return nil
	}

	for k, v := range tags {
		codebuildTags = append(codebuildTags, &codebuild.Tag{Key: aws.String(k), Value: aws.String(v)})
	}

	return codebuildTags
}

func convertTaskStatus(codebuildStatus string) string {
	switch codebuildStatus {
	case codebuild.StatusTypeStopped:
		return launcher.TaskStopped
	case codebuild.StatusTypeInProgress:
		return launcher.TaskRunning
	case codebuild.StatusTypeSucceeded:
		return launcher.TaskSucceeded
	default:
		return launcher.TaskFailed
	}
}
