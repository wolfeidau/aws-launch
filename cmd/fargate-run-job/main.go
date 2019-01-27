package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/onrik/logrus/filename"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wolfeidau/fargate-run-job/pkg/launcher"
)

var (
	app     = kingpin.New("fargate-run-job", "A command-line fargate provisioning application.")
	verbose = app.Flag("verbose", "Verbose mode.").Short('v').Bool()

	newDef  = app.Command("new-definition", "Build a new definition for an ECS task.")
	defFile = newDef.Arg("def-file", "The path to the definition file.").Required().File()

	newTask    = app.Command("new-task", "Launch a new ECS task.")
	launchFile = newTask.Arg("launch-file", "The path to the launch parameters file.").Required().File()
)

func main() {

	logrus.AddHook(filename.NewHook())
	config := aws.NewConfig()

	lch := launcher.NewECSLauncher(config)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case newDef.FullCommand():
		ld := new(launcher.DefinitionParams)

		err := loadJSONFile(*defFile, ld)
		if err != nil {
			logrus.WithError(err).Fatal("failed to load definition file")
		}

		logrus.WithFields(logrus.Fields{
			"name": ld.ECS.DefinitionName,
		}).Info("new definition")

		defTag, err := lch.CreateDefinition(ld)
		if err != nil {
			logrus.Fatalf("failed to create definition: %v", err)
		}

		logrus.WithField("defTag", defTag).Info("created")

	case newTask.FullCommand():

		rt := new(launcher.RunTaskParams)

		err := loadJSONFile(*launchFile, rt)
		if err != nil {
			logrus.WithError(err).Fatal("failed to load definition file")
		}

		logrus.WithFields(logrus.Fields{
			"name": rt.ServiceName,
		}).Info("new task")

		err = lch.RunTask(rt)
		if err != nil {
			logrus.WithError(err).Fatal("failed to launch task")
		}
	}
}

func loadJSONFile(file *os.File, val interface{}) error {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}

	err = json.Unmarshal(data, val)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal JSON file")
	}

	return nil
}

func emptyToNil(val *string) *string {
	if val == nil {
		return val
	}
	if *val == "" {
		return nil
	}
	return val
}
