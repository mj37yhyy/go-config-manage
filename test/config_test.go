package test

import (
	"fmt"
	"github.com/mj37yhyy/go-config-manage"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	v := viper.New()
	if err := v.AddRemoteProvider("consul", "localhost:8500", "go-config"); err != nil {
		log.Println(err)
		return
	}
	v.SetConfigType("yaml") // Need to explicitly set this to json
	if err := v.ReadRemoteConfig(); err != nil {
		log.Println(err)
		return
	}
	if err := v.Unmarshal(&conf); err != nil {
		log.Println(err)
		return
	}
	v2 := viper.New()
	if err := v2.AddRemoteProvider("consul", "localhost:8500", "a/b/c"); err != nil {
		log.Println(err)
		return
	}
	v2.SetConfigType("yaml") // Need to explicitly set this to json
	if err := v2.ReadRemoteConfig(); err != nil {
		log.Println(err)
		return
	}
	if err := v2.Unmarshal(&conf); err != nil {
		log.Println(err)
		return
	}
	log.Println(conf)
}
