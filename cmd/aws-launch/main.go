package main

import (
	"fmt"
	"os"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
	"github.com/wolfeidau/aws-launch/pkg/configuration"
	"github.com/wolfeidau/aws-launch/pkg/launcher"
	"github.com/wolfeidau/aws-launch/pkg/launcher/service"
)

var (
	app     = kingpin.New("aws-launch", "A command-line task provisioning application.")
	verbose = app.Flag("verbose", "Verbose mode.").Short('v').Bool()

	oneTask = app.Command("one-task", "Create a new definition and run in one shot.")
	oneFile = oneTask.Arg("one-file", "The path to the definition and run file.").Required().File()

	defineTask = app.Command("define-task", "Create a new definition.")
	defFile    = defineTask.Arg("def-file", "The path to the definition file.").Required().File()

	launchTask = app.Command("launch-task", "Launch a new task.")
	launchFile = launchTask.Arg("launch-file", "The path to the launch parameters file.").Required().File()

	cleanupTask = app.Command("cleanup-task", "Cleanup a new task.")
	cleanupFile = cleanupTask.Arg("cleanup-file", "The path to the cleanup parameters file.").Required().File()

	getTaskLogs     = app.Command("get-task-logs", "Get logs for task.")
	getTaskLogsFile = getTaskLogs.Arg("get-task-logs", "The path to the get task logs parameters file.").Required().File()

	dumpSchema = app.Command("dump-schema", "Write the JSON Schema to stdout.")
	structName = dumpSchema.Arg("struct-name", "The name of the struct you want to retrieve the schema.").Required().Enum("DefineAndLaunchParams", "DefineTaskParams", "LaunchTaskParams")
)

func main() {

	logrus.AddHook(filename.NewHook())
	config := aws.NewConfig()

	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	lch := service.New(config)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case oneTask.FullCommand():

		dlp := new(launcher.DefineAndLaunchParams)

		data, err := configuration.LoadJSONFile(*oneFile, dlp)
		if err != nil {
			logrus.WithError(err).Fatal("failed to load definition file")
		}

		err = configuration.ValidateInputFile("DefineAndLaunchParams", string(data))
		if err != nil {
			logrus.WithError(err).Fatal("failed to load definition file")
		}

		logrus.Info("valid task supplied")

		logrus.Info("new task")

		res, err := lch.DefineAndLaunch(dlp)
		if err != nil {
			logrus.WithError(err).Fatal("failed to launch task")
		}

		rt := dlp.BuildLaunchTask(res.ID)

		waitRes, err := lch.WaitForTask(&launcher.WaitForTaskParams{
			ID:        res.ID,
			ECS:       rt.ECS,
			Codebuild: rt.Codebuild,
		})
		if err != nil {
			logrus.WithError(err).Fatal("failed to wait for task")
		}

		getRes, err := lch.GetTaskStatus(&launcher.GetTaskStatusParams{
			ID:        waitRes.ID,
			ECS:       rt.ECS,
			Codebuild: rt.Codebuild,
		})
		if err != nil {
			logrus.WithError(err).Fatal("failed to get task status")
		}

		elapsed := getRes.EndTime.Sub(*getRes.StartTime)

		logrus.WithFields(logrus.Fields{
			"ID":                     getRes.ID,
			"TaskStatus":             getRes.TaskStatus,
			"DefinitionID":           res.DefinitionID,
			"CloudwatchLogGroupName": res.CloudwatchLogGroupName,
			"Elapsed":                elapsed,
		}).Info("run task complete")

	case defineTask.FullCommand():
		ld := new(launcher.DefineTaskParams)

		data, err := configuration.LoadJSONFile(*defFile, ld)
		if err != nil {
			logrus.WithError(err).Fatal("failed to load definition file")
		}

		err = configuration.ValidateInputFile("DefinitionParams", string(data))
		if err != nil {
			logrus.WithError(err).Fatal("failed to load definition file")
		}

		logrus.Info("valid definition supplied")

		logrus.Info("new definition")

		defTag, err := lch.DefineTask(ld)
		if err != nil {
			logrus.Fatalf("failed to create definition: %v", err)
		}

		logrus.WithField("ID", defTag.ID).Info("created")

	case launchTask.FullCommand():

		rt := new(launcher.LaunchTaskParams)

		data, err := configuration.LoadJSONFile(*launchFile, rt)
		if err != nil {
			logrus.WithError(err).Fatal("failed to load definition file")
		}

		err = configuration.ValidateInputFile("LaunchTaskParams", string(data))
		if err != nil {
			logrus.WithError(err).Fatal("failed to load definition file")
		}

		logrus.Info("valid task supplied")

		logrus.Info("new task")

		res, err := lch.LaunchTask(rt)
		if err != nil {
			logrus.WithError(err).Fatal("failed to launch task")
		}

		waitRes, err := lch.WaitForTask(&launcher.WaitForTaskParams{
			ID:        res.ID,
			ECS:       rt.ECS,
			Codebuild: rt.Codebuild,
		})
		if err != nil {
			logrus.WithError(err).Fatal("failed to wait for task")
		}

		getRes, err := lch.GetTaskStatus(&launcher.GetTaskStatusParams{
			ID:        waitRes.ID,
			ECS:       rt.ECS,
			Codebuild: rt.Codebuild,
		})
		if err != nil {
			logrus.WithError(err).Fatal("failed to get task status")
		}

		elapsed := getRes.EndTime.Sub(*getRes.StartTime)

		logrus.WithFields(logrus.Fields{
			"ID":      getRes.ID,
			"Elapsed": elapsed,
		}).Info("run task complete")

	case cleanupTask.FullCommand():

		ctp := new(launcher.CleanupTaskParams)

		data, err := configuration.LoadJSONFile(*cleanupFile, ctp)
		if err != nil {
			logrus.WithError(err).Fatal("failed to load definition file")
		}

		err = configuration.ValidateInputFile("CleanupTaskParams", string(data))
		if err != nil {
			logrus.WithError(err).Fatal("failed to load definition file")
		}

		logrus.Info("valid task supplied")

		logrus.Info("cleanup task")

		_, err = lch.CleanupTask(ctp)
		if err != nil {
			logrus.WithError(err).Fatal("failed to cleanup task")
		}

		logrus.Info("cleanup task complete")

	case getTaskLogs.FullCommand():

		gtlp := new(launcher.GetTaskLogsParams)

		data, err := configuration.LoadJSONFile(*getTaskLogsFile, gtlp)
		if err != nil {
			logrus.WithError(err).Fatal("failed to load definition file")
		}

		err = configuration.ValidateInputFile("GetTaskLogsParams", string(data))
		if err != nil {
			logrus.WithError(err).Fatal("failed to load definition file")
		}

		logrus.Info("valid task supplied")

		logrus.Info("get task logs")

		getTaskLogsRes, err := lch.GetTaskLogs(gtlp)
		if err != nil {
			logrus.WithError(err).Fatal("failed to cleanup task")
		}

		logrus.WithField("NextToken", aws.StringValue(getTaskLogsRes.NextToken)).Info("GetTaskLogs result")

		for _, logLine := range getTaskLogsRes.LogLines {
			fmt.Printf("%s::%s\n", logLine.Timestamp.Format(time.RFC3339Nano), logLine.Message)
		}

	case dumpSchema.FullCommand():

		jsonStr, err := configuration.GetSchema(*structName)
		if err != nil {
			logrus.WithError(err).Fatal("failed to marshal schema")
		}

		fmt.Println(jsonStr)
	}
}
