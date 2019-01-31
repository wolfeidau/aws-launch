package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/onrik/logrus/filename"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wolfeidau/fargate-run-job/pkg/launcher"
	"github.com/wolfeidau/fargate-run-job/pkg/schema"
	"github.com/xeipuuv/gojsonschema"
)

var (
	app     = kingpin.New("fargate-run-job", "A command-line fargate provisioning application.")
	verbose = app.Flag("verbose", "Verbose mode.").Short('v').Bool()

	newDef  = app.Command("new-definition", "Build a new definition.")
	defFile = newDef.Arg("def-file", "The path to the definition file.").Required().File()

	newTask    = app.Command("new-task", "Launch a new task.")
	launchFile = newTask.Arg("launch-file", "The path to the launch parameters file.").Required().File()

	dumpSchema = app.Command("dump-schema", "Write the JSON Schema to stdout.")
	structName = dumpSchema.Arg("struct-name", "The name of the struct you want to retrieve the schema.").Required().Enum("DefinitionParams", "RunTaskParams")
)

func main() {

	logrus.AddHook(filename.NewHook())
	config := aws.NewConfig()

	lch := launcher.New(config)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case newDef.FullCommand():
		ld := new(launcher.DefinitionParams)

		data, err := loadJSONFile(*defFile, ld)
		if err != nil {
			logrus.WithError(err).Fatal("failed to load definition file")
		}

		err = validateInputFile("DefinitionParams", string(data))
		if err != nil {
			logrus.WithError(err).Fatal("failed to load definition file")
		}

		logrus.Info("valid definition supplied")

		logrus.Info("new definition")

		defTag, err := lch.CreateDefinition(ld)
		if err != nil {
			logrus.Fatalf("failed to create definition: %v", err)
		}

		logrus.WithField("ID", defTag.ID).Info("created")

	case newTask.FullCommand():

		rt := new(launcher.RunTaskParams)

		data, err := loadJSONFile(*launchFile, rt)
		if err != nil {
			logrus.WithError(err).Fatal("failed to load definition file")
		}

		err = validateInputFile("RunTaskParams", string(data))
		if err != nil {
			logrus.WithError(err).Fatal("failed to load definition file")
		}

		logrus.Info("valid task supplied")

		logrus.Info("new task")

		res, err := lch.RunTask(rt)
		if err != nil {
			logrus.WithError(err).Fatal("failed to launch task")
		}

		elapsed := res.EndTime.Sub(*res.StartTime)

		logrus.WithFields(logrus.Fields{
			"ID": res.ID,
			"Elapsed": fmt.Sprintf("%s", elapsed),
		}).Info("run task complete")

	case dumpSchema.FullCommand():

		jsonStr, err := getSchema(*structName)
		if err != nil {
			logrus.WithError(err).Fatal("failed to marshal schema")
		}

		fmt.Println(jsonStr)
	}
}

func validateInputFile(paramName string, payloadJSON string) error {

	jsonStr, err := getSchema(paramName)
	if err != nil {
		return err
	}

	schemaLoader := gojsonschema.NewStringLoader(jsonStr)
	documentLoader := gojsonschema.NewStringLoader(payloadJSON)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		for _, desc := range result.Errors() {
			logrus.WithField("desc", desc).Warn("validation")
		}

		return errors.New("validation errors")
	}

	return nil
}

func getSchema(paramName string) (string, error) {
	var v interface{}

	switch paramName {
	case "DefinitionParams":
		v = &launcher.DefinitionParams{}
	case "RunTaskParams":
		v = &launcher.RunTaskParams{}
	}

	data, err := schema.DumpSchema(v)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func loadJSONFile(file *os.File, val interface{}) ([]byte, error) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	err = json.Unmarshal(data, val)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal JSON file")
	}

	return data, nil
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
