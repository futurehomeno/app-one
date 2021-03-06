package router

import (
	"fmt"
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/app-one/model"
	"path/filepath"
	"strings"
	"time"
)

type FromFimpRouter struct {
	inboundMsgCh fimpgo.MessageCh
	mqt          *fimpgo.MqttTransport
	instanceId   string
	appLifecycle *model.Lifecycle
	configs      *model.Configs
	hideFlag     bool
}

func NewFromFimpRouter(mqt *fimpgo.MqttTransport,appLifecycle *model.Lifecycle,configs *model.Configs) *FromFimpRouter {
	fc := FromFimpRouter{inboundMsgCh: make(fimpgo.MessageCh,5),mqt:mqt,appLifecycle:appLifecycle,configs:configs}
	fc.hideFlag = false
	fc.mqt.RegisterChannel("ch1",fc.inboundMsgCh)
	return &fc
}

func (fc *FromFimpRouter) Start() {

	// TODO: Choose either adapter or app topic

	// ------ Adapter topics ---------------------------------------------
	fc.mqt.Subscribe(fmt.Sprintf("pt:j1/+/rt:dev/rn:%s/ad:1/#",model.ServiceName))
	fc.mqt.Subscribe(fmt.Sprintf("pt:j1/+/rt:ad/rn:%s/ad:1",model.ServiceName))

	// ------ Application topic -------------------------------------------
	//fc.mqt.Subscribe(fmt.Sprintf("pt:j1/+/rt:app/rn:%s/ad:1",model.ServiceName))

	go func(msgChan fimpgo.MessageCh) {
		for  {
			select {
			case newMsg :=<- msgChan:
				fc.routeFimpMessage(newMsg)
			}
		}
	}(fc.inboundMsgCh)
}

func (fc *FromFimpRouter) routeFimpMessage(newMsg *fimpgo.Message) {
	log.Debug("New fimp msg")
	addr := strings.Replace(newMsg.Addr.ServiceAddress,"_0","",1)
	switch newMsg.Payload.Service {
	case "out_lvl_switch" :
		addr = strings.Replace(addr,"l","",1)
		switch newMsg.Payload.Type {
		case "cmd.binary.set":
			// TODO: This is example . Add your logic here or remove
		case "cmd.lvl.set":
			// TODO: This is an example . Add your logic here or remove
		}
	case "out_bin_switch":
		log.Debug("Sending switch")
		// TODO: This is an example . Add your logic here or remove
	case model.ServiceName:
		adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: model.ServiceName, ResourceAddress:"1"}
		switch newMsg.Payload.Type {
		case "cmd.auth.login":
			authReq := model.Login{}
			err := newMsg.Payload.GetObjectValue(&authReq)
			if err != nil {
				log.Error("Incorrect login message ")
				return
			}
			status := model.AuthStatus{
				Status:    model.AuthStateAuthenticated,
				ErrorText: "",
				ErrorCode: "",
			}
			if authReq.Username != "" && authReq.Password != ""{
				// TODO: This is an example . Add your logic here or remove
			}else {
				status.Status = "ERROR"
				status.ErrorText = "Empty username or password"
			}
			fc.appLifecycle.SetAuthState(model.AuthStateAuthenticated)
			msg := fimpgo.NewMessage("evt.auth.status_report",model.ServiceName,fimpgo.VTypeObject,status,nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
				// if response topic is not set , sending back to default application event topic
				log.Info("NEW TOKENS username = %s , password = %s",authReq.Username,authReq.Password)
				fc.mqt.Publish(adr,msg)
			}

		case "cmd.auth.logout":
			status := model.AuthStatus{
				Status:    model.AuthStateNotAuthenticated,
				ErrorText: "",
				ErrorCode: "",
			}
			fc.appLifecycle.SetAuthState(model.AuthStateNotAuthenticated)
			msg := fimpgo.NewMessage("evt.auth.status_report",model.ServiceName,fimpgo.VTypeObject,status,nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
				// if response topic is not set , sending back to default application event topic
				fc.mqt.Publish(adr,msg)
			}

		case "cmd.auth.set_tokens":
			authReq := model.SetTokens{}
			err := newMsg.Payload.GetObjectValue(&authReq)
			if err != nil {
				log.Error("Incorrect login message ")
				return
			}
			status := model.AuthStatus{
				Status:    model.AuthStateAuthenticated,
				ErrorText: "",
				ErrorCode: "",
			}
			if authReq.AccessToken != "" && authReq.RefreshToken != ""{
				// TODO: This is an example . Add your logic here or remove
				log.Info("NEW TOKENS access_token = %s , refresh_token = %s",authReq.AccessToken,authReq.RefreshToken)
			}else {
				status.Status = "ERROR"
				status.ErrorText = "Empty username or password"
			}

			fc.appLifecycle.SetAuthState(model.AuthStateAuthenticated)
			msg := fimpgo.NewMessage("evt.auth.status_report",model.ServiceName,fimpgo.VTypeObject,status,nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
				// if response topic is not set , sending back to default application event topic
				fc.mqt.Publish(adr,msg)
			}

		case "cmd.app.get_manifest":
			mode,err := newMsg.Payload.GetStringValue()
			if err != nil {
				log.Error("Incorrect request format ")
				return
			}
			manifest := model.NewManifest()
			err = manifest.LoadFromFile(filepath.Join(fc.configs.GetDefaultDir(),"app-manifest.json"))
			if err != nil {
				log.Error("Failed to load manifest file .Error :",err.Error())
				return
			}
			fc.configs.Param4 = time.Now().Format(time.RFC3339)
			fc.configs.Param5 = time.Now().Format(time.Kitchen)+"  🤖  📡"
			if mode == "manifest_state" {
				manifest.AppState = *fc.appLifecycle.GetAllStates()
				if fc.configs.Param6 == 0 {
					fc.configs.Param6 = 19
				}
				manifest.ConfigState = fc.configs
			}

			if uiBlock := manifest.GetUIBlock("security");uiBlock != nil {
				uiBlock.Hidden = fc.hideFlag
			}

			if uiButton := manifest.GetButton("factory_reset");uiButton != nil {
				uiButton.Hidden = fc.hideFlag
			}

			if uiConfig := manifest.GetAppConfig("param_1");uiConfig != nil {
				uiConfig.Hidden = fc.hideFlag
			}

			msg := fimpgo.NewMessage("evt.app.manifest_report",model.ServiceName,fimpgo.VTypeObject,manifest,nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
				// if response topic is not set , sending back to default application event topic
				fc.mqt.Publish(adr,msg)
			}
		case "cmd.app.show_hide":
			if fc.hideFlag {
				fc.hideFlag = false
			}else {
				fc.hideFlag = true
			}
			val := model.ButtonActionResponse{
				Operation:       "cmd.app.show_hide",
				OperationStatus: "ok",
				Next:            "config",
				ErrorCode:       "",
				ErrorText:       "",
			}
			msg := fimpgo.NewMessage("evt.app.config_action_report",model.ServiceName,fimpgo.VTypeObject,val,nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
				fc.mqt.Publish(adr,msg)
			}

		case "cmd.app.get_state":
			msg := fimpgo.NewMessage("evt.app.manifest_report",model.ServiceName,fimpgo.VTypeObject,fc.appLifecycle.GetAllStates(),nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
				// if response topic is not set , sending back to default application event topic
				fc.mqt.Publish(adr,msg)
			}

		case "cmd.config.get_extended_report":

			msg := fimpgo.NewMessage("evt.config.extended_report",model.ServiceName,fimpgo.VTypeObject,fc.configs,nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
				fc.mqt.Publish(adr,msg)
			}

		case "cmd.config.extended_set":
			conf := model.Configs{}
			err :=newMsg.Payload.GetObjectValue(&conf)
			if err != nil {
				// TODO: This is an example . Add your logic here or remove
				log.Error("Can't parse configuration object")
				return
			}
			fc.configs.Param1 = conf.Param1
			fc.configs.Param2 = conf.Param2
			fc.configs.Param3 = conf.Param3
			fc.configs.Param4 = conf.Param4
			fc.configs.Param6 = conf.Param6
			if conf.AuthType != "" {
				fc.configs.AuthType = conf.AuthType
				manifest := model.NewManifest()
				manifestFile := filepath.Join(fc.configs.GetDefaultDir(),"app-manifest.json")
				err = manifest.LoadFromFile(manifestFile)
				if err != nil {
					log.Error("Failed to load manifest file .Error :",err.Error())
					return
				}
				manifest.Auth.Type = conf.AuthType
				manifest.SaveToFile(manifestFile)
			}
			fc.configs.SaveToFile()

			logLevel, err := log.ParseLevel(conf.LogLevel)
			if err == nil {
				log.SetLevel(logLevel)
				fc.configs.LogLevel = conf.LogLevel
			}

			log.Debugf("App reconfigured . New parameters : %v",fc.configs)
			// TODO: This is an example . Add your logic here or remove
			configReport := model.ConfigReport{
				OpStatus: "ok",
				AppState:  *fc.appLifecycle.GetAllStates(),
			}
			fc.appLifecycle.SetConfigState(model.ConfigStateConfigured)
			fc.appLifecycle.SetAppState(model.AppStateRunning,nil)
			msg := fimpgo.NewMessage("evt.app.config_report",model.ServiceName,fimpgo.VTypeObject,configReport,nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
				fc.mqt.Publish(adr,msg)
			}

		case "cmd.log.set_level":
			// Configure log level
			level , err :=newMsg.Payload.GetStringValue()
			if err != nil {
				return
			}
			logLevel, err := log.ParseLevel(level)
			if err == nil {
				log.SetLevel(logLevel)
				fc.configs.LogLevel = level
				fc.configs.SaveToFile()
			}
			log.Info("Log level updated to = ",logLevel)

		case "cmd.app.factory_reset":
			val := model.ButtonActionResponse{
				Operation:       "cmd.app.factory_reset",
				OperationStatus: "ok",
				Next:            "config",
				ErrorCode:       "",
				ErrorText:       "",
			}
			fc.appLifecycle.SetConfigState(model.ConfigStateNotConfigured)
			fc.appLifecycle.SetAppState(model.AppStateNotConfigured,nil)
			fc.appLifecycle.SetAuthState(model.AuthStateNotAuthenticated)
			msg := fimpgo.NewMessage("evt.app.config_action_report",model.ServiceName,fimpgo.VTypeObject,val,nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
				fc.mqt.Publish(adr,msg)
			}

		case "cmd.system.reconnect":
			// This is optional operation.
			fc.appLifecycle.PublishEvent(model.EventConfigured,"from-fimp-router",nil)
			val := model.ButtonActionResponse{
				Operation:       "cmd.system.reconnect",
				OperationStatus: "ok",
				Next:            "config",
				ErrorCode:       "",
				ErrorText:       "",
			}
			msg := fimpgo.NewMessage("evt.app.config_action_report",model.ServiceName,fimpgo.VTypeObject,val,nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
				fc.mqt.Publish(adr,msg)
			}

		case "cmd.network.get_all_nodes":
			// TODO: This is an example . Add your logic here or remove
		case "cmd.thing.get_inclusion_report":
			//nodeId , _ := newMsg.Payload.GetStringValue()
			// TODO: This is an example . Add your logic here or remove
		case "cmd.thing.inclusion":
			//flag , _ := newMsg.Payload.GetBoolValue()
			// TODO: This is an example . Add your logic here or remove
		case "cmd.thing.delete":
			// remove device from network
			val,err := newMsg.Payload.GetStrMapValue()
			if err != nil {
				log.Error("Wrong msg format")
				return
			}
			deviceId , ok := val["address"]
			if ok {
				// TODO: This is an example . Add your logic here or remove
				log.Info(deviceId)
			}else {
				log.Error("Incorrect address")

			}
		}

	}

}


