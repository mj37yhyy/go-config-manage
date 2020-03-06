package test

import (
	"fmt"
	"github.com/mj37yhyy/go-config-manage"
	"testing"
)

type TApplication struct {
	Name string `mapstructure:"name"`
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
	fmt.Println("==========================", conf.Spring.Application.Name)
}
