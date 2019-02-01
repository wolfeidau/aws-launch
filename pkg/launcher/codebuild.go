package launcher

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
)

// CodeBuildLauncher used to launch containers in CodeBuild
type CodeBuildLauncher struct {
	codeBuildSvc codebuildiface.CodeBuildAPI
	cwlogsSvc    cloudwatchlogsiface.CloudWatchLogsAPI
}

// NewCodeBuildLauncher create a new launcher
func NewCodeBuildLauncher(cfgs ...*aws.Config) *CodeBuildLauncher {
	sess := session.New(cfgs...)
	return &CodeBuildLauncher{
		codeBuildSvc: codebuild.New(sess),
		cwlogsSvc:    cloudwatchlogs.New(sess),
	}
}

// CreateDefinition create a codebuild job for this definition and return the ARN of this job
func (cbl *CodeBuildLauncher) CreateDefinition(dp *DefinitionParams) (*CreateDefinitionResult, error) {

	logGroupName := fmt.Sprintf("/aws/codebuild/%s", dp.Codebuild.ProjectName)

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
	}

	res, err := cbl.codeBuildSvc.CreateProject(&codebuild.CreateProjectInput{
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
	if err != nil {
		return nil, errors.Wrap(err, "failed to register project.")
	}

	return &CreateDefinitionResult{
		ID: aws.StringValue(res.Project.Arn),
	}, nil
}

// RunTask run a container task and monitor it till completion
func (cbl *CodeBuildLauncher) RunTask(rt *RunTaskParams) (*RunTaskResult, error) {

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

	taskRes :=  &BaseTaskResult{
		ID: aws.StringValue(res.Build.Id),
	}

	return &RunTaskResult{taskRes}, nil
}

// WaitForTask wait for task to complete
func (cbl *CodeBuildLauncher) WaitForTask(wft *WaitForTaskParams) (*WaitForTaskResult, error) {

	params := &codebuild.BatchGetBuildsInput{
		Ids: []*string{aws.String(wft.ID)},
	}

	err := cbl.waitUntilTasksStoppedWithContext(context.Background(), params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start build.")
	}

	return &WaitForTaskResult{ID: wft.ID}, nil
}

// GetTaskStatus get task status
func (cbl *CodeBuildLauncher) GetTaskStatus(gts *GetTaskStatusParams) (*GetTaskStatusResult, error) {
	
	params := &codebuild.BatchGetBuildsInput{
		Ids: []*string{aws.String(gts.ID)},
	}
	getBuildRes, err := cbl.codeBuildSvc.BatchGetBuilds(params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start build.")
	}

	logrus.WithFields(logrus.Fields{
		"BuildComplete": aws.BoolValue(getBuildRes.Builds[0].BuildComplete),
		"BuildStatus":   aws.StringValue(getBuildRes.Builds[0].BuildStatus),
		"StartTime":     aws.TimeValue(getBuildRes.Builds[0].StartTime),
		"StopTime":      aws.TimeValue(getBuildRes.Builds[0].EndTime),
	}).Info("Describe completed Task")

	taskRes := &BaseTaskResult{
		ID: aws.StringValue(getBuildRes.Builds[0].Arn),
		StartTime: getBuildRes.Builds[0].StartTime,
		EndTime: getBuildRes.Builds[0].EndTime,
		CodeBuild: &RunTaskCodebuildResult{
			BuildArn: aws.StringValue(getBuildRes.Builds[0].Arn),
			BuildStatus: aws.StringValue(getBuildRes.Builds[0].BuildStatus),
		},
		Successful: false,
	}

	if aws.StringValue(getBuildRes.Builds[0].BuildStatus) == "SUCCEEDED" {
		taskRes.Successful = false
	}

	return &GetTaskStatusResult{taskRes}, nil

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
