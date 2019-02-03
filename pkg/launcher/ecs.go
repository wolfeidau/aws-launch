package launcher

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultCPU default cpu allocation to a fargate task
	DefaultCPU = "256"
	// DefaultMemory default memory allocation to a fargate task
	DefaultMemory = "512"
)

// ECSLauncher used to launch containers in ECS, specifically fargate
type ECSLauncher struct {
	ecsSvc    ecsiface.ECSAPI
	cwlogsSvc cloudwatchlogsiface.CloudWatchLogsAPI
}

// NewECSLauncher create a new launcher
func NewECSLauncher(cfgs ...*aws.Config) *ECSLauncher {
	sess := session.New(cfgs...)
	return &ECSLauncher{
		ecsSvc:    ecs.New(sess),
		cwlogsSvc: cloudwatchlogs.New(sess),
	}
}

// DefineAndLaunch define and launch a container in ECS
func (lc *ECSLauncher) DefineAndLaunch(dlp *DefineAndLaunchParams) (*DefineAndLaunchResult, error) {
	defRes, err := lc.DefineTask(dlp.BuildDefineTask())
	if err != nil {
		return nil, errors.Wrap(err, "define failed.")
	}

	launchRes, err := lc.LaunchTask(dlp.BuildLaunchTask(defRes.ID))
	if err != nil {
		return nil, errors.Wrap(err, "launch failed.")
	}

	return &DefineAndLaunchResult{
		BaseTaskResult:         launchRes.BaseTaskResult,
		CloudwatchLogGroupName: defRes.CloudwatchLogGroupName,
		DefinitionID:           defRes.ID,
	}, nil
}

// DefineTask create a container task definition
func (lc *ECSLauncher) DefineTask(dp *DefineTaskParams) (*DefineTaskResult, error) {

	logGroupName := fmt.Sprintf("/ecs/fargate/%s", dp.ECS.DefinitionName)

	_, err := lc.cwlogsSvc.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
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

	// register the task definition with default base memory, cpu and cwlogs groups
	res, err := lc.ecsSvc.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
		RequiresCompatibilities: aws.StringSlice([]string{
			"FARGATE",
		}),
		Family:      aws.String(dp.ECS.DefinitionName),
		TaskRoleArn: dp.TaskRoleARN,
		NetworkMode: aws.String(ecs.NetworkModeAwsvpc),
		Cpu:         aws.String(DefaultCPU),
		Memory:      aws.String(DefaultMemory),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			&ecs.ContainerDefinition{
				Name:  aws.String(dp.ECS.ContainerName),
				Image: aws.String(dp.Image),
				LogConfiguration: &ecs.LogConfiguration{
					LogDriver: aws.String(ecs.LogDriverAwslogs),
					Options: map[string]*string{
						"awslogs-group":         aws.String(logGroupName),
						"awslogs-region":        aws.String(dp.Region),
						"awslogs-stream-prefix": aws.String("ecs"),
					},
				},
				Environment: convertMapToKeyValuePair(dp.Environment),
			},
		},
		ExecutionRoleArn: aws.String(dp.ECS.ExecutionRoleARN),
		Tags:             convertMapToECSTags(dp.Tags),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to register task definition.")
	}

	logrus.WithField("result", res).Debug("Register Task Definition")

	return &DefineTaskResult{
		ID:                     fmt.Sprintf("%s:%d", aws.StringValue(res.TaskDefinition.Family), aws.Int64Value(res.TaskDefinition.Revision)),
		CloudwatchLogGroupName: logGroupName,
	}, nil
}

// LaunchTask run a container task
func (lc *ECSLauncher) LaunchTask(lp *LaunchTaskParams) (*LaunchTaskResult, error) {

	logrus.WithFields(logrus.Fields{
		"ClusterName":    lp.ECS.ClusterName,
		"TaskDefinition": lp.ECS.TaskDefinition,
	}).Info("Launch Task")

	runRes, err := lc.ecsSvc.RunTask(&ecs.RunTaskInput{
		Cluster:        aws.String(lp.ECS.ClusterName),
		LaunchType:     aws.String(ecs.LaunchTypeFargate),
		TaskDefinition: aws.String(lp.ECS.TaskDefinition),
		Count:          aws.Int64(1),
		Overrides: &ecs.TaskOverride{
			ContainerOverrides: []*ecs.ContainerOverride{
				&ecs.ContainerOverride{
					Cpu:         aws.Int64(lp.ECS.CPU),
					Memory:      aws.Int64(lp.ECS.Memory),
					Name:        aws.String(lp.ECS.ContainerName),
					Environment: convertMapToKeyValuePair(lp.Environment),
				},
			},
		},
		PlatformVersion: aws.String("LATEST"),
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(ecs.AssignPublicIpEnabled),
				Subnets:        aws.StringSlice(lp.ECS.Subnets),
			},
		},
		Tags: convertMapToECSTags(lp.Tags),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create task.")
	}

	logrus.WithFields(logrus.Fields{
		"TaskID": shortenTaskArn(runRes.Tasks[0].TaskArn),
	}).Info("Task Provisioned")

	taskRes := &BaseTaskResult{
		ID: aws.StringValue(runRes.Tasks[0].TaskArn),
	}

	return &LaunchTaskResult{taskRes}, nil
}

// WaitForTask wait for task to complete
func (lc *ECSLauncher) WaitForTask(wft *WaitForTaskParams) (*WaitForTaskResult, error) {

	descInput := &ecs.DescribeTasksInput{
		Cluster: aws.String(wft.ECS.ClusterName),
		Tasks:   []*string{aws.String(wft.ID)},
	}

	err := lc.ecsSvc.WaitUntilTasksStopped(descInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check stopped task.")
	}

	return &WaitForTaskResult{ID: wft.ID}, nil
}

// GetTaskStatus get task status
func (lc *ECSLauncher) GetTaskStatus(gts *GetTaskStatusParams) (*GetTaskStatusResult, error) {
	descInput := &ecs.DescribeTasksInput{
		Cluster: aws.String(gts.ECS.ClusterName),
		Tasks:   []*string{aws.String(gts.ID)},
	}
	descRes, err := lc.ecsSvc.DescribeTasks(descInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to describe task.")
	}

	logrus.WithFields(logrus.Fields{
		"TaskID":        shortenTaskArn(descRes.Tasks[0].TaskArn),
		"StopCode":      aws.StringValue(descRes.Tasks[0].StopCode),
		"StoppedReason": aws.StringValue(descRes.Tasks[0].StoppedReason),
	}).Info("Describe completed Task")

	taskRes := &BaseTaskResult{
		ID:         aws.StringValue(descRes.Tasks[0].TaskArn),
		StartTime:  descRes.Tasks[0].StartedAt,
		EndTime:    descRes.Tasks[0].StoppedAt,
		TaskStatus: TaskRunning,
		ECS: &LaunchTaskECSResult{
			TaskArn: aws.StringValue(descRes.Tasks[0].TaskArn),
			TaskID:  shortenTaskArn(descRes.Tasks[0].TaskArn),
		},
	}

	if aws.StringValue(descRes.Tasks[0].LastStatus) == "STOPPED" {
		if aws.StringValue(descRes.Tasks[0].StopCode) == "EssentialContainerExited" {
			taskRes.TaskStatus = TaskSucceeded
		} else {
			taskRes.TaskStatus = TaskFailed
		}
	}

	return &GetTaskStatusResult{taskRes}, nil
}

// CleanupTask clean up ecs task definition
func (lc *ECSLauncher) CleanupTask(ctp *CleanupTaskParams) (*CleanupTaskResult, error) {
	_, err := lc.ecsSvc.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: aws.String(ctp.ECS.TaskDefinition),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to de-register definition.")
	}

	return &CleanupTaskResult{}, nil
}

func shortenTaskArn(taskArn *string) string {
	tokens := strings.Split(aws.StringValue(taskArn), "/")
	if len(tokens) == 3 {
		return tokens[2]
	}

	return "unknown"
}

func convertMapToKeyValuePair(env map[string]string) []*ecs.KeyValuePair {

	ecsEnv := []*ecs.KeyValuePair{}

	// empty map is valid
	if env == nil {
		return nil
	}

	for k, v := range env {
		ecsEnv = append(ecsEnv, &ecs.KeyValuePair{Name: aws.String(k), Value: aws.String(v)})
	}

	return ecsEnv
}

func convertMapToECSTags(tags map[string]string) []*ecs.Tag {

	ecsTags := []*ecs.Tag{}

	// empty map is valid
	if tags == nil {
		return nil
	}

	for k, v := range tags {
		ecsTags = append(ecsTags, &ecs.Tag{Key: aws.String(k), Value: aws.String(v)})
	}

	return ecsTags
}
