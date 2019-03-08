package main

import (
	"flag"
	"fmt"
	"maxpanda-alarm/config"
	"maxpanda-alarm/core"
	"maxpanda-alarm/ierror"
	"net/http"

	"github.com/bitly/go-simplejson"
	sms "github.com/fwhappy/aliyun-sms/app"
	"github.com/fwhappy/mail/exmail"
	"github.com/fwhappy/util"
)

var (
	// 环境，用来读取不同的配置文件
	env = flag.String("env", "local", "env")
	// 本机内网ip
	host = flag.String("host", "127.0.0.1", "host")
	// 监听端口
	port = flag.Int("port", 13579, "port")
)

func init() {
	// 解析url参数
	flag.Parse()
}

func main() {
	defer util.RecoverPanic()
	// 初始化日志配置
	core.LoadLoggerConfig(core.GetConfigFile("log.toml", *env, "conf/env"))
	defer core.Logger.Flush()

	http.HandleFunc("/test", hello)
	// 报警(手机+邮件)
	http.HandleFunc("/alarm", alarm)
	// 报警邮件
	http.HandleFunc("/alarm/mail", mailAlarm)
	// 手机报警
	http.HandleFunc("/alarm/mobile", mobileAlarm)

	err := http.ListenAndServe(fmt.Sprintf("%v:%v", *host, *port), nil)
	if err != nil {
		core.Logger.Error("ListenAndServe:%v", err)
		return
	}
}

func getRequestParam(r *http.Request) (*simplejson.Json, error) {
	r.ParseForm()
	var content string
	content = r.PostFormValue("*")
	if content == "" {
		content = r.Form.Get("*")
	}
	core.Logger.Debug("content:%v", content)
	j, err := simplejson.NewJson([]byte(content))
	if err != nil {
		core.Logger.Error("[getRequestParam]解析参数失败, content:%v", content)
	}
	return j, err
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

func alarm(w http.ResponseWriter, r *http.Request) {
	param, err := getRequestParam(r)
	if err != nil {
		return
	}
	// 发送邮件报警
	mailErr := sendMailAlarm(param)
	// 发送短信报警
	mobileErr := sendMobileAlarm(param)
	if mailErr != nil || mobileErr != nil {
		core.Logger.Info("[alarm]fail, mailErr:%v, mobileErr:%v", mailErr, mobileErr)
	} else {
		core.Logger.Info("[alarm]success")
	}
	w.Write([]byte("ok"))
}

func mailAlarm(w http.ResponseWriter, r *http.Request) {
	param, err := getRequestParam(r)
	if err != nil {
		return
	}
	// 发送邮件报警
	sendMailAlarm(param)
	w.Write([]byte("ok"))
}

func mobileAlarm(w http.ResponseWriter, r *http.Request) {
	param, err := getRequestParam(r)
	if err != nil {
		return
	}
	// 发送短信报警
	sendMobileAlarm(param)
	w.Write([]byte("ok"))
}

// 读取配置文件
func sendMailAlarm(params *simplejson.Json) *ierror.Error {
	// 检查参数完整性
	alarmTarget, err := params.Get("target").String()
	if err != nil {
		return ierror.NewError(-100, "sendMailAlarm", "target")
	}
	// 检查是否支持alarmType
	if !util.InStringSlice(alarmTarget, config.EnableAlarmTarget) {
		return ierror.NewError(-102, alarmTarget)
	}
	cfgFile := core.GetConfigFile(alarmTarget+".toml", "alarm", "conf/share")
	alarmConfig, err := core.LoadConfig(cfgFile)
	if err != nil {
		return ierror.NewError(-103, cfgFile)
	}
	subject, err := params.Get("subject").String()
	if err != nil {
		return ierror.NewError(-100, "sendMailAlarm", "subject")
	}
	body, err := params.Get("body").String()
	if err != nil {
		return ierror.NewError(-100, "sendMailAlarm", "body")
	}
	// 异步发送
	go func() {
		host := alarmConfig["mail_host"].(string)
		port := int(alarmConfig["mail_port"].(int64))
		email := alarmConfig["mail_from_address"].(string)
		password := alarmConfig["mail_from_password"].(string)
		addressList := alarmConfig["mail_alarm_address_list"].(string)
		from := alarmConfig["mail_from_name"].(string)
		// 发送邮件
		sendErrors := exmail.Send(host, port, email, password, addressList, from, subject, body)
		if len(sendErrors) > 0 {
			for _, err := range sendErrors {
				core.Logger.Error("[sendMailAlarm]Send Mail error:", err.Error())
			}
		}
	}()
	return nil
}

// 手机
func sendMobileAlarm(params *simplejson.Json) error {
	// 检查参数完整性
	alarmTarget, err := params.Get("target").String()
	if err != nil {
		return ierror.NewError(-100, "sendMobileAlarm", "target")
	}
	// 检查是否支持alarmType
	if !util.InStringSlice(alarmTarget, config.EnableAlarmTarget) {
		return ierror.NewError(-102, alarmTarget)
	}
	cfgFile := core.GetConfigFile(alarmTarget+".toml", "alarm", "conf/share")
	alarmConfig, err := core.LoadConfig(cfgFile)
	if err != nil {
		return ierror.NewError(-103, cfgFile)
	}
	smsCode, err := params.Get("sms_code").String()
	if err != nil {
		return ierror.NewError(-100, "sendMobileAlarm", "sms_code")
	}
	smsParams, _ := params.Get("sms_params").String()
	core.Logger.Debug("smsParams:%v", smsParams)
	// 异步发送
	go func() {
		accessKeyId := alarmConfig["sms_access_key_id"].(string)
		accessKeySecret := alarmConfig["sms_access_key_secret"].(string)
		phoneNumbers := alarmConfig["sms_number_list"].(string)
		signName := alarmConfig["sms_access_key_sign_name"].(string)

		smsClient := sms.NewSmsClient("http://dysmsapi.aliyuncs.com/")
		if _, err := smsClient.Execute(accessKeyId, accessKeySecret, phoneNumbers, signName, smsCode, smsParams); err != nil {
			core.Logger.Error("[sendMobileAlarm]Send Mobile error:", err.Error())
		}
	}()
	return nil
}
