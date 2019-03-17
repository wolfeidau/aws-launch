package ecs

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
	"github.com/wolfeidau/aws-launch/pkg/cwlogs"
	"github.com/wolfeidau/aws-launch/pkg/launcher"
)

const (
	// DefaultCPU default cpu allocation to a fargate task
	DefaultCPU = "256"

	// DefaultMemory default memory allocation to a fargate task
	DefaultMemory = "512"

	// ECSStreamPrefix the prefix used in the ECS cloudwatch log stream name
	ECSStreamPrefix = "ecs"

	// ECSLogGroupFormat the name format for ECS cloudwatch log group names
	ECSLogGroupFormat = "/aws/fargate/%s"
)

// ECSLauncher used to launch containers in ECS, specifically fargate
type ECSLauncher struct {
	ecsSvc       ecsiface.ECSAPI
	cwlogsSvc    cloudwatchlogsiface.CloudWatchLogsAPI
	cwlogsReader cwlogs.LogsReader
}

// NewECSLauncher create a new launcher
func NewECSLauncher(cfgs ...*aws.Config) *ECSLauncher {
	sess := session.Must(session.NewSession(cfgs...))
	return &ECSLauncher{
		ecsSvc:       ecs.New(sess),
		cwlogsSvc:    cloudwatchlogs.New(sess),
		cwlogsReader: cwlogs.NewCloudwatchLogsReader(cfgs...),
	}
}

// DefineTask create a container task definition
func (lc *ECSLauncher) DefineTask(dp *launcher.DefineTaskParams) (*launcher.DefineTaskResult, error) {

	logGroupName := fmt.Sprintf(ECSLogGroupFormat, dp.ECS.DefinitionName)

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

		logrus.WithField("name", logGroupName).Info("cloudwatch log group exists")
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
			{
				Name:  aws.String(dp.ECS.ContainerName),
				Image: aws.String(dp.Image),
				LogConfiguration: &ecs.LogConfiguration{
					LogDriver: aws.String(ecs.LogDriverAwslogs),
					Options: map[string]*string{
						"awslogs-group":         aws.String(logGroupName),
						"awslogs-region":        aws.String(dp.Region),
						"awslogs-stream-prefix": aws.String(ECSStreamPrefix),
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

	return &launcher.DefineTaskResult{
		ID:                     fmt.Sprintf("%s:%d", aws.StringValue(res.TaskDefinition.Family), aws.Int64Value(res.TaskDefinition.Revision)),
		CloudwatchLogGroupName: logGroupName,
		CloudwatchStreamPrefix: "ecs",
	}, nil
}

// LaunchTask run a container task
func (lc *ECSLauncher) LaunchTask(lp *launcher.LaunchTaskParams) (*launcher.LaunchTaskResult, error) {

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
				{
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

	task := runRes.Tasks[0]

	logrus.WithFields(logrus.Fields{
		"TaskID": shortenTaskArn(task.TaskArn),
	}).Info("Task Provisioned")

	taskRes := &launcher.BaseTaskResult{
		ID:         aws.StringValue(task.TaskArn),
		TaskStatus: launcher.TaskRunning,
		ECS: &launcher.LaunchTaskECSResult{
			TaskArn: aws.StringValue(task.TaskArn),
			TaskID:  shortenTaskArn(task.TaskArn),
		},
	}

	return &launcher.LaunchTaskResult{BaseTaskResult: taskRes}, nil
}

// WaitForTask wait for task to complete
func (lc *ECSLauncher) WaitForTask(wft *launcher.WaitForTaskParams) (*launcher.WaitForTaskResult, error) {

	descInput := &ecs.DescribeTasksInput{
		Cluster: aws.String(wft.ECS.ClusterName),
		Tasks:   []*string{aws.String(wft.ID)},
	}

	err := lc.ecsSvc.WaitUntilTasksStopped(descInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check stopped task.")
	}

	return &launcher.WaitForTaskResult{ID: wft.ID}, nil
}

// GetTaskStatus get task status
func (lc *ECSLauncher) GetTaskStatus(gts *launcher.GetTaskStatusParams) (*launcher.GetTaskStatusResult, error) {
	descInput := &ecs.DescribeTasksInput{
		Cluster: aws.String(gts.ECS.ClusterName),
		Tasks:   []*string{aws.String(gts.ID)},
	}
	descRes, err := lc.ecsSvc.DescribeTasks(descInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to describe task.")
	}

	task := descRes.Tasks[0]

	logrus.WithFields(logrus.Fields{
		"TaskID":        shortenTaskArn(task.TaskArn),
		"StopCode":      aws.StringValue(task.StopCode),
		"StoppedReason": aws.StringValue(task.StoppedReason),
	}).Info("Describe completed Task")

	taskRes := &launcher.BaseTaskResult{
		ID:         aws.StringValue(task.TaskArn),
		StartTime:  task.StartedAt,
		EndTime:    task.StoppedAt,
		TaskStatus: launcher.TaskRunning,
		ECS: &launcher.LaunchTaskECSResult{
			TaskArn: aws.StringValue(task.TaskArn),
			TaskID:  shortenTaskArn(task.TaskArn),
		},
	}

	if aws.StringValue(task.LastStatus) == "STOPPED" {
		if aws.StringValue(task.StopCode) == "EssentialContainerExited" {
			taskRes.TaskStatus = launcher.TaskSucceeded
		} else {
			taskRes.TaskStatus = launcher.TaskFailed
		}
	}

	return &launcher.GetTaskStatusResult{BaseTaskResult: taskRes}, nil
}

// CleanupTask clean up ecs task definition
func (lc *ECSLauncher) CleanupTask(ctp *launcher.CleanupTaskParams) (*launcher.CleanupTaskResult, error) {
	_, err := lc.ecsSvc.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: aws.String(ctp.ECS.TaskDefinition),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to de-register definition.")
	}

	return &launcher.CleanupTaskResult{}, nil
}

// GetTaskLogs get task logs
func (lc *ECSLauncher) GetTaskLogs(gtlp *launcher.GetTaskLogsParams) (*launcher.GetTaskLogsResult, error) {
	taskID := shortenTaskArn(aws.String(gtlp.ECS.TaskARN))
	logGroupName := fmt.Sprintf(ECSLogGroupFormat, gtlp.ECS.DefinitionName)
	streamName := fmt.Sprintf("%s/%s/%s", ECSStreamPrefix, gtlp.ECS.DefinitionName, taskID)

	logrus.WithFields(logrus.Fields{
		"group":  logGroupName,
		"stream": streamName,
	}).Info("ReadLogs")

	res, err := lc.cwlogsReader.ReadLogs(&cwlogs.ReadLogsParams{
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
