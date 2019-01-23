package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("starting")

	res, err := http.Get("https://accounts.google.com/.well-known/openid-configuration")
	if err != nil {
		log.Fatalf("failed to retrieve openid wellknown endpoint: %+v", err)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("failed to read body of openid wellknown JSON: %+v", err)
	}

	wellknown := make(map[string]interface{})

	err = json.Unmarshal(data, &wellknown)
	if err != nil {
		log.Fatalf("failed to read body of openid wellknown JSON: %+v", err)
	}

	logrus.WithField("openid-configuration", wellknown).Info("JSON")
}
