package model

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/app-one/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const ServiceName  = "app-one"

type Configs struct {
	path                  string
	InstanceAddress       string `json:"instance_address"`
	MqttServerURI         string `json:"mqtt_server_uri"`
	MqttUsername          string `json:"mqtt_server_username"`
	MqttPassword          string `json:"mqtt_server_password"`
	MqttClientIdPrefix    string `json:"mqtt_client_id_prefix"`
	LogFile               string `json:"log_file"`
	LogLevel              string `json:"log_level"`
	LogFormat             string `json:"log_format"`
	WorkDir               string `json:"-"`
	ConfiguredAt          string `json:"configured_at"`
	ConfiguredBy          string `json:"configured_by"`
	Param1                bool   `json:"param_1"`
	Param2                string `json:"param_2"`
	Param3                []string `json:"param_3"`
	Param4                string `json:"param_4"`
	Param5                string `json:"param_5"`
	Param6                int    `json:"param_6"`
	AuthType              string `json:"auth_type"`
}

func NewConfigs(workDir string) *Configs {
	conf := &Configs{WorkDir: workDir}
	conf.path = filepath.Join(workDir,"data","config.json")
	if !utils.FileExists(conf.path) {
		log.Info("Config file doesn't exist.Loading default config")
		defaultConfigFile := filepath.Join(workDir,"defaults","config.json")
		err := utils.CopyFile(defaultConfigFile,conf.path)
		if err != nil {
			fmt.Print(err)
			panic("Can't copy config file.")
		}
	}
	return conf
}

func (cf * Configs) LoadFromFile() error {
	configFileBody, err := ioutil.ReadFile(cf.path)
	if err != nil {
		cf.InitDefault()
		return cf.SaveToFile()
	}
	err = json.Unmarshal(configFileBody, cf)
	if err != nil {
		return err
	}
	return nil
}

func (cf *Configs) SaveToFile() error {
	cf.ConfiguredBy = "auto"
	cf.ConfiguredAt = time.Now().Format(time.RFC3339)
	bpayload, err := json.Marshal(cf)
	err = ioutil.WriteFile(cf.path, bpayload, 0664)
	if err != nil {
		return err
	}
	return err
}

func (cf *Configs) GetDataDir()string {
	return filepath.Join(cf.WorkDir,"data")
}

func (cf *Configs) GetDefaultDir()string {
	return filepath.Join(cf.WorkDir,"defaults")
}

func (cf * Configs) LoadDefaults()error {
	configFile := filepath.Join(cf.WorkDir,"data","config.json")
	os.Remove(configFile)
	log.Info("Config file doesn't exist.Loading default config")
	defaultConfigFile := filepath.Join(cf.WorkDir,"defaults","config.json")
	return utils.CopyFile(defaultConfigFile,configFile)
}

func (cf *Configs) InitDefault() {
	cf.InstanceAddress = "1"
	cf.MqttServerURI = "tcp://localhost:1883"
	cf.MqttClientIdPrefix = "app-one"
	cf.LogFile = "/var/log/thingsplex/app-one/app-one.log"
	cf.WorkDir = "/opt/thingsplex/app-one"
	cf.LogLevel = "debug"
	cf.LogFormat = "text"
	cf.Param1 = true
	cf.Param2 = "test"
	cf.Param4 = "ASDAQE-ADSFDA"
	cf.Param6 = 19
	cf.AuthType = "password"
}

func (cf *Configs) IsConfigured()bool {
	// TODO : Add logic here
	return true
}

type ConfigReport struct {
	OpStatus string `json:"op_status"`
	AppState AppStates `json:"app_state"`
}