package test

import (
	"fmt"
	"github.com/mj37yhyy/go-config-manage"
	log "github.com/sirupsen/logrus"
	_ "github.com/spf13/viper/remote"
	"gopkg.in/yaml.v2"
	"reflect"
	"testing"
	"time"
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

	//fmt.Println("==========================", conf)
	for {
		fmt.Println("==========================", conf)
		time.Sleep(time.Second * 2) // 每次请求后延迟
	}
}

func Test2(t *testing.T) {
	var conf = TRoot{}
	test4(&conf)
	log.Info(conf)
}
func test4(out interface{}) {
	test3(out)
}

func test3(out interface{}) {
	/*targetActual := reflect.ValueOf(out).Elem()
	log.Info("targetActual",targetActual)
	configType := targetActual.Type()
	log.Info("configType",configType)
	baseReflect := reflect.New(configType)
	log.Info("baseReflect",baseReflect)
	// Actual type.
	base := baseReflect.Interface()
	log.Info("base",base)
	*/
	b2 := reflect.New(reflect.ValueOf(out).Elem().Type()).Interface()

	/*v := reflect.ValueOf(b2)
	log.Info(v.Kind())
	log.Info(v.Elem())*/
	var str = `spring:
  application:
    name: myApp
    version: v1.0.0`
	yaml.Unmarshal([]byte(str), b2)

	reflect.ValueOf(out).Elem().Set(reflect.ValueOf(b2).Elem())
	//out = reflect.ValueOf(b2).Elem()
	//log.Info(out)
}
