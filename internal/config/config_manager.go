package config

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"brandlovrs.exporter/pkg/util"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/config"
)

func getResourceFolder() string {
	f, _ := os.Getwd()
	return fmt.Sprintf("%s/%s", filepath.Dir(f), filepath.Base(f))
}

func init() {
	viper.SetConfigFile(getResourceFolder() + "/.env")
	err := viper.ReadInConfig()

	if err != nil {
		return
	}

	watcherConfig()
}

func ReadParameter(parameter string) string {
	return fmt.Sprint(viper.Get(parameter))
}

func watcherConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		err := viper.ReadInConfig()
		if err != nil {
			util.Sugar.Errorf("An error ocurred")
		}
	})
}

// A struct to represent the configuration of a self-contained unit of your
// application.
type cfg struct {
	Parameter string
}

func LoadParameter(key string) string {

	environment := os.Getenv("PIPES_ENV")

	var (
		content []byte
		err     error
	)

	switch environment {
	case "dev":
		content, err = os.ReadFile(getResourceFolder() + "/application-dev.yaml")
	case "prod":
		content, err = os.ReadFile(getResourceFolder() + "/application-prod.yaml")
	default:
		content, err = os.ReadFile(getResourceFolder() + "/application-dev.yaml")
	}

	if err != nil {
		log.Fatal(err)
	}

	var c cfg
	provider, err := config.NewYAML(config.Source(bytes.NewReader(content)))

	err2 := provider.Get(key).Populate(&c)
	if err != nil || err2 != nil {
		util.Sugar.Infof(err.Error())
		util.Sugar.Infof(err2.Error())
		return ""
	}

	return c.Parameter
}
