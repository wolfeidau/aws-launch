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

// Launcher used to launch containers in ECS, specifically fargate
type Launcher struct {
	ecsSvc       ecsiface.ECSAPI
	cwlogsSvc    cloudwatchlogsiface.CloudWatchLogsAPI
	cwlogsReader cwlogs.LogsReader
}

// NewLauncher create a new launcher
func NewLauncher(cfgs ...*aws.Config) LauncherAPI {
	sess := session.Must(session.NewSession(cfgs...))
	return &Launcher{
		ecsSvc:       ecs.New(sess),
		cwlogsSvc:    cloudwatchlogs.New(sess),
		cwlogsReader: cwlogs.NewCloudwatchLogsReader(cfgs...),
	}
}

// DefineTask create a container task definition
func (lc *Launcher) DefineTask(dp *DefineTaskParams) (*DefineTaskResult, error) {

	logGroupName := fmt.Sprintf(ECSLogGroupFormat, dp.DefinitionName)

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
		Family:      aws.String(dp.DefinitionName),
		TaskRoleArn: dp.TaskRoleARN,
		NetworkMode: aws.String(ecs.NetworkModeAwsvpc),
		Cpu:         aws.String(DefaultCPU),
		Memory:      aws.String(DefaultMemory),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Name:  aws.String(dp.ContainerName),
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
		ExecutionRoleArn: aws.String(dp.ExecutionRoleARN),
		Tags:             convertMapToECSTags(dp.Tags),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to register task definition.")
	}

	logrus.WithField("result", res).Debug("Register Task Definition")

	return &DefineTaskResult{
		ID:                     fmt.Sprintf("%s:%d", aws.StringValue(res.TaskDefinition.Family), aws.Int64Value(res.TaskDefinition.Revision)),
		CloudwatchLogGroupName: logGroupName,
		CloudwatchStreamPrefix: "ecs",
	}, nil
}

// LaunchTask run a container task
func (lc *Launcher) LaunchTask(lp *LaunchTaskParams) (*LaunchTaskResult, error) {

	logrus.WithFields(logrus.Fields{
		"ClusterName":    lp.ClusterName,
		"TaskDefinition": lp.TaskDefinition,
	}).Info("Launch Task")

	runRes, err := lc.ecsSvc.RunTask(&ecs.RunTaskInput{
		Cluster:        aws.String(lp.ClusterName),
		LaunchType:     aws.String(ecs.LaunchTypeFargate),
		TaskDefinition: aws.String(lp.TaskDefinition),
		Count:          aws.Int64(1),
		Overrides: &ecs.TaskOverride{
			ContainerOverrides: []*ecs.ContainerOverride{
				{
					Cpu:         aws.Int64(lp.CPU),
					Memory:      aws.Int64(lp.Memory),
					Name:        aws.String(lp.ContainerName),
					Environment: convertMapToKeyValuePair(lp.Environment),
				},
			},
		},
		PlatformVersion: aws.String("LATEST"),
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: aws.String(ecs.AssignPublicIpEnabled),
				Subnets:        aws.StringSlice(lp.Subnets),
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

	taskRes := &LaunchTaskResult{
		ID:         aws.StringValue(task.TaskArn),
		TaskStatus: launcher.TaskRunning,
		TaskArn:    aws.StringValue(task.TaskArn),
		TaskID:     shortenTaskArn(task.TaskArn),
	}

	return taskRes, nil
}

// WaitForTask wait for task to complete
func (lc *Launcher) WaitForTask(wft *WaitForTaskParams) (*WaitForTaskResult, error) {

	descInput := &ecs.DescribeTasksInput{
		Cluster: aws.String(wft.ClusterName),
		Tasks:   []*string{aws.String(wft.ID)},
	}

	err := lc.ecsSvc.WaitUntilTasksStopped(descInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check stopped task.")
	}

	return &WaitForTaskResult{ID: wft.ID}, nil
}

// GetTaskStatus get task status
func (lc *Launcher) GetTaskStatus(gts *GetTaskStatusParams) (*GetTaskStatusResult, error) {
	descInput := &ecs.DescribeTasksInput{
		Cluster: aws.String(gts.ClusterName),
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

	taskRes := &GetTaskStatusResult{
		ID:         aws.StringValue(task.TaskArn),
		StartTime:  task.StartedAt,
		EndTime:    task.StoppedAt,
		TaskStatus: convertTaskStatus(aws.StringValue(task.LastStatus), aws.StringValue(task.StopCode)),
		TaskArn:    aws.StringValue(task.TaskArn),
		TaskID:     shortenTaskArn(task.TaskArn),
		LastStatus: aws.StringValue(task.LastStatus),
		StopCode:   aws.StringValue(task.StopCode),
	}

	return taskRes, nil
}

// StopTask clean up ecs task definition
func (lc *Launcher) StopTask(stp *StopTaskParams) (*StopTaskResult, error) {
	res, err := lc.ecsSvc.StopTask(&ecs.StopTaskInput{
		Cluster: aws.String(stp.ClusterName),
		Reason:  aws.String("request stop task"),
		Task:    aws.String(stp.TaskARN),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to de-register definition.")
	}

	task := res.Task

	return &StopTaskResult{
		LastStatus: aws.StringValue(task.LastStatus),
		StopCode:   aws.StringValue(task.StopCode),
		TaskStatus: convertTaskStatus(aws.StringValue(task.LastStatus), aws.StringValue(task.StopCode)),
	}, nil
}

// CleanupTask clean up ecs task definition
func (lc *Launcher) CleanupTask(ctp *CleanupTaskParams) (*CleanupTaskResult, error) {
	_, err := lc.ecsSvc.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: aws.String(ctp.TaskDefinition),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to de-register definition.")
	}

	return &CleanupTaskResult{}, nil
}

// GetTaskLogs get task logs
func (lc *Launcher) GetTaskLogs(gtlp *GetTaskLogsParams) (*GetTaskLogsResult, error) {
	taskID := shortenTaskArn(aws.String(gtlp.TaskARN))
	logGroupName := fmt.Sprintf(ECSLogGroupFormat, gtlp.DefinitionName)
	streamName := fmt.Sprintf("%s/%s/%s", ECSStreamPrefix, gtlp.DefinitionName, taskID)

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

	return &GetTaskLogsResult{
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

func convertTaskStatus(lastStatus, stopCode string) string {
	if lastStatus == ecs.DesiredStatusStopped {
		if stopCode == ecs.TaskStopCodeEssentialContainerExited {
			return launcher.TaskSucceeded
		}
		return launcher.TaskFailed
	}
	return launcher.TaskRunning
}
