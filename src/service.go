package main

import (
	"flag"
	"fmt"
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/discovery"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/app-one/model"
	"github.com/thingsplex/app-one/router"
	"github.com/thingsplex/app-one/utils"
	"time"
)


func main() {
	var workDir string
	flag.StringVar(&workDir, "c", "", "Work dir")
	flag.Parse()
	if workDir == "" {
		workDir = "./"
	} else {
		fmt.Println("Work dir ", workDir)
	}
	appLifecycle := model.NewAppLifecycle()
	configs := model.NewConfigs(workDir)
	err := configs.LoadFromFile()
	if err != nil {
		fmt.Print(err)
		panic("Can't load config file.")
	}

	utils.SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
	log.Info("--------------Starting app-one----------------")
	log.Info("Work directory : ",configs.WorkDir)
	appLifecycle.PublishEvent(model.EventConfiguring, "main", nil)

	mqtt := fimpgo.NewMqttTransport(configs.MqttServerURI, configs.MqttClientIdPrefix, configs.MqttUsername, configs.MqttPassword, true, 1, 1)
	err = mqtt.Start()
	responder := discovery.NewServiceDiscoveryResponder(mqtt)
	responder.RegisterResource(model.GetDiscoveryResource())
	responder.Start()

	fimpRouter := router.NewFromFimpRouter(mqtt,appLifecycle,configs)
	fimpRouter.Start()

	//------------------ Sample code --------------------------------------

	msg := fimpgo.NewFloatMessage("evt.sensor.report", "temp_sensor", float64(35.5), nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "app-one", ResourceAddress: "1", ServiceName: "temp_sensor", ServiceAddress: "300"}
	mqtt.Publish(&adr, msg)
	if err != nil {
		log.Error("Can't connect to broker. Error:", err.Error())
	} else {
		log.Info("Connected")
	}
	appLifecycle.SetAppState(model.AppStateRunning,nil)

	//------------------ Sample code --------------------------------------

	for {
		appLifecycle.WaitForState("main", model.AppStateRunning)
		// Configure custom resources here
		//if err := conFimpRouter.Start(); err !=nil {
		//	appLifecycle.PublishEvent(model.EventConfigError,"main",nil)
		//}else {
		//	appLifecycle.WaitForState(model.StateConfiguring,"main")
		//}
		//TODO: Add logic here
		appLifecycle.WaitForState(model.AppStateNotConfigured,"main")
	}

	mqtt.Stop()
	time.Sleep(5 * time.Second)
}
