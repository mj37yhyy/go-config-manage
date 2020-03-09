package test

import (
	"fmt"
	consulapi "github.com/armon/consul-api"
	"github.com/mj37yhyy/go-config-manage"
	log "github.com/sirupsen/logrus"
	_ "github.com/spf13/viper/remote"
	"gopkg.in/yaml.v2"
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
	var machines = []string{"localhost:8500"}
	c, err := newConsulClient(machines)
	if err != nil {
		panic(err)
	}
	b, err2 := get4ConsulKV(c, "a/b/c", "")
	if err2 != nil {
		panic(err2)
	}
	if err := yaml.Unmarshal(b, &conf); err != nil {
		panic(err)
	}
	b2, err := yaml.Marshal(conf)
	if err != nil {
		panic(err)
	}
	log.Info(string(b2))

}
func newConsulClient(machines []string) (*consulapi.KV, error) {
	conf := consulapi.DefaultConfig()
	if len(machines) > 0 {
		conf.Address = machines[0]
	}
	client, err := consulapi.NewClient(conf)
	if err != nil {
		return nil, err
	}
	return client.KV(), nil
}

func get4ConsulKV(kvClient *consulapi.KV, key string, token string) ([]byte, error) {
	kv, _, err := kvClient.Get(key, &consulapi.QueryOptions{
		Token: token,
	})
	if err != nil {
		return nil, err
	}
	if kv == nil {
		return nil, fmt.Errorf("Key ( %s ) was not found.", key)
	}
	return kv.Value, nil
}
