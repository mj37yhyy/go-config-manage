package test

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/mitchellh/mapstructure"
	"github.com/mj37yhyy/go-config-manage"
	log "github.com/sirupsen/logrus"
	_ "github.com/spf13/viper/remote"
	"os"
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
	log.Info(conf)
	log.Info(&conf)
	test5(&conf)
	//log.Info(conf)
}

func TestB(t *testing.T) {

	_path, err := os.Getwd()
	fmt.Println(_path, err)
}

func TestA(t *testing.T) {
	//var jsonStr = `{"A1":1,"B1":2,"C1":"aaa","D1":"ddfdf","E1":"asdfasdf"}`
	var yamlStr = `A1: 1
B1: 2
C1: aaa
D1: asdfsad
E1: 1111`
	type A struct {
		A1 int `yaml:"A1"`
		B1 int `yaml:"B1"`
	}

	type B struct {
		C1 string `yaml:"C1"`
		D1 string `yaml:"D1"`
	}

	type C struct {
		A
		B
		E1 string `yaml:"E1"`
	}

	c := C{}

	/*if err := json.Unmarshal([]byte(jsonStr), &c); err != nil {
		panic(err)
	}*/

	if err := yaml.Unmarshal([]byte(yamlStr), &c); err != nil {
		panic(err)
	}
	fmt.Println(c)
	fmt.Println(c.A)
}
func test5(conf *TRoot) {
	log.Info(conf)
	log.Info(&conf)
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

type ServerConfig struct {
	Id      string `json:"Id"`
	Address string `json:"Address"`
	Version string `json:"Version"`
}

func TestQueryDistrbute(t *testing.T) {
	var j = `[{"Id":"1","Address":"1","Version":"1"},{"Id":"2","Address":"2","Version":"2"}]`
	var data interface{}
	if err := json.Unmarshal([]byte(j), &data); err != nil {
		panic(err)
	}

	var serverConfig []ServerConfig
	if value, ok := data.([]interface{}); ok {
		if err := mapstructure.Decode(value, &serverConfig); err != nil {
			log.Errorf("err %v", err)
		}
	}
	log.Info(serverConfig)
}
