package test

import (
	"fmt"
	"github.com/mj37yhyy/go-config-manage"
	log "github.com/sirupsen/logrus"
	_ "github.com/spf13/viper/remote"
	"testing"
)

type TApplication struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type TSpring struct {
	Application TApplication `mapstructure:"application"`
}

type TRoot struct {
	Spring TSpring `mapstructure:"spring"`
}

func TestInitConfig(t *testing.T) {
	var conf = TRoot{}
	if err := go_config_manage.InitConfig(&conf); err != nil {
		panic(err)
	}

	fmt.Println("==========================", conf)
	/*for i := 0; i < 5; i = 0 {
		fmt.Println("==========================", conf)
		time.Sleep(time.Second * 2) // 每次请求后延迟
	}*/
}

func Test2(t *testing.T) {
	var conf = TRoot{}
	var applicationConfigPath = "applicationConfigPath"
	var applicationConfigName = "applicationConfigPath"
	var applicationConfigType = "applicationConfigPath"
	log.WithFields(log.Fields{
		"root":                  conf,
		"applicationConfigPath": applicationConfigPath,
		"applicationConfigName": applicationConfigName,
		"applicationConfigType": applicationConfigType,
	}).Info("func initBootstrap end")
}
