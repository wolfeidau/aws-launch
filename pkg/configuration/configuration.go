package configuration

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wolfeidau/fargate-run-job/pkg/launcher"
	"github.com/wolfeidau/fargate-run-job/pkg/schema"
	"github.com/xeipuuv/gojsonschema"
)

// LoadJSONFile load a JSON configuration file
func LoadJSONFile(file *os.File, val interface{}) ([]byte, error) {
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

// ValidateInputFile load a JSON configuration file
func ValidateInputFile(paramName string, payloadJSON string) error {

	jsonStr, err := GetSchema(paramName)
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

// GetSchema get the schema definition for the supplied parameter by name
func GetSchema(paramName string) (string, error) {
	var v interface{}

	switch paramName {
	case "DefineAndLaunchParams":
		v = &launcher.DefineAndLaunchParams{}
	case "DefinitionParams":
		v = &launcher.DefineTaskParams{}
	case "LaunchTaskParams":
		v = &launcher.LaunchTaskParams{}
	}

	data, err := schema.DumpSchema(v)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
