package go_config_manage

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	consulapi "github.com/armon/consul-api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	FlagBootstrapConfigPath string = "BootstrapConfigPath"
	FlagBootstrapConfigName string = "BootstrapConfigName"
	FlagBootstrapConfigType string = "BootstrapConfigType"

	FlagApplicationConfigPath string = "ApplicationConfigPath"
	FlagApplicationConfigName string = "ApplicationConfigName"
	FlagApplicationConfigType string = "ApplicationConfigType"

	EnvBootstrapConfigPath = FlagBootstrapConfigPath
	EnvBootstrapConfigName = FlagBootstrapConfigName
	EnvBootstrapConfigType = FlagBootstrapConfigType

	EnvApplicationConfigPath = FlagApplicationConfigPath
	EnvApplicationConfigName = FlagApplicationConfigName
	EnvApplicationConfigType = FlagApplicationConfigType

	bootstrapName string = "bootstrap"
	configName    string = "application"
	configType    string = "yaml"
)

type Root struct {
	Application Application `mapstructure:"application"`
}

type Application struct {
	Config Config `mapstructure:"config"`
}

type Config struct {
	Remote []Remote `mapstructure:"remote"`
	File   bool     `mapstructure:"file"`
}

type Remote struct {
	Enabled  bool     `mapstructure:"enabled"`
	Format   string   `mapstructure:"format"`
	Provider string   `mapstructure:"provider"`
	Endpoint []string `mapstructure:"endpoint"`
	Path     []string `mapstructure:"path"`
	Token    string   `mapstructure:"token"`
}

// 入口函数
func InitConfig(obj interface{}) error {
	log.Trace("func InitConfig begin")

	// 初始化引导配置
	root, applicationConfigPath, applicationConfigName, applicationConfigType, err := initBootstrap()
	if err != nil {
		log.WithField("err", err).Error("func initBootstrap error")
		return err
	}

	// 初始化应用配置
	err = initApplication(root, applicationConfigPath, applicationConfigName, applicationConfigType, obj)
	if err != nil {
		log.WithField("err", err).Error("func initApplication error")
		return err
	}
	log.Trace("func InitConfig end")
	return nil
}

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

// 初始化用户配置
func initApplication(root Root,
	applicationConfigPath string,
	applicationConfigName string,
	applicationConfigType string,
	obj interface{}) error {
	log.WithFields(log.Fields{
		"root":                  root,
		"applicationConfigPath": applicationConfigPath,
		"applicationConfigName": applicationConfigName,
		"applicationConfigType": applicationConfigType,
	}).Trace("func initApplication begin")
	if root.Application.Config.File {
		log.Trace("从文件中获取")
		if applicationConfigPath != "" && applicationConfigName != "" && applicationConfigType != "" {
			// 如果传入的application位置存在
			log.Trace("传入的application位置存在")
			if err := newConfig(obj, applicationConfigPath, applicationConfigName, applicationConfigType); err != nil {
				log.WithField("err", err).Error("func initApplication error")
				return err
			}
		} else {
			log.Trace("从默认位置读取")
			if err := newConfig(obj, "", configName, configType); err != nil {
				log.WithField("err", err).Error("func initApplication error")
				return err
			}
		}
	}

	// 从远端取
	for _, remote := range root.Application.Config.Remote {
		log.WithFields(log.Fields{
			"remote": remote,
		}).Trace("从远程获取")
		if remote.Enabled {
			kvClient, err := newConsulClient(remote.Endpoint)
			if err != nil {
				log.WithField("err", err).Error("newConsulClient error")
				return err
			}
			getKV(remote, kvClient, obj)
			go func() {
				for {
					getKV(remote, kvClient, obj)
					time.Sleep(time.Second * 5) // 每次请求后延迟
				}
			}()
		}
	}
	return nil
}

// 从远端获取
func getKV(remote Remote, kvClient *consulapi.KV, obj interface{}) {
	newObj := reflect.New(reflect.ValueOf(obj).Elem().Type()).Interface()
	for _, path := range remote.Path {
		b, err := get4ConsulKV(kvClient, path, remote.Token)
		if err != nil {
			log.Errorf("unable to read remote config: %v", err)
			continue
		}
		in := bytes.NewReader(b)
		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(in); err != nil {
			log.Errorf("buf.ReadFrom error: %v", err)
			continue
		}
		if remote.Format != "" {
			if strings.EqualFold(remote.Format, "yaml") {
				if err := yaml.Unmarshal(buf.Bytes(), newObj); err != nil {
					log.Errorf("unmarshal yaml config error: %v", err)
					continue
				}
			} else if strings.EqualFold(remote.Format, "json") {
				if err := json.Unmarshal(buf.Bytes(), newObj); err != nil {
					log.Errorf("unmarshal json config error: %v", err)
					continue
				}
			}
		}
	}
	reflect.ValueOf(obj).Elem().Set(reflect.ValueOf(newObj).Elem())
}

// 初始化引导文件
func initBootstrap() (Root, string, string, string, error) {
	log.Trace("func initBootstrap begin")
	var root = Root{}
	var applicationConfigPath, applicationConfigName, applicationConfigType string
	// 从命令行获取
	flagBootstrapConfigPath, flagBootstrapConfigName, flagBootstrapConfigType,
		flagApplicationConfigPath, flagApplicationConfigName, flagApplicationConfigType,
		err := getFlag()

	log.WithFields(log.Fields{
		"flagBootstrapConfigPath":   flagBootstrapConfigPath,
		"flagBootstrapConfigName":   flagBootstrapConfigName,
		"flagBootstrapConfigType":   flagBootstrapConfigType,
		"flagApplicationConfigPath": flagApplicationConfigPath,
		"flagApplicationConfigName": flagApplicationConfigName,
		"flagApplicationConfigType": flagApplicationConfigType,
	}).Trace("从命令行获取")

	if err != nil {
		log.WithField("err", err).Error("func initBootstrap error")
		return Root{}, "", "", "", err
	}
	if flagBootstrapConfigPath != "" && flagBootstrapConfigName != "" && flagBootstrapConfigType != "" {
		// application文件位置
		log.Trace("从用户定义的位置读取application文件")
		applicationConfigPath, applicationConfigName, applicationConfigType =
			flagApplicationConfigPath, flagApplicationConfigName, flagApplicationConfigType
		if err := newConfig(&root, flagBootstrapConfigPath, flagBootstrapConfigName, flagBootstrapConfigType); err != nil {
			log.WithField("err", err).Error("func initBootstrap error")
			return Root{}, "", "", "", err
		}
	} else {
		// 从env获取
		envBootstrapConfigPath, envBootstrapConfigName, envBootstrapConfigType,
			envApplicationConfigPath, envApplicationConfigName, envApplicationConfigType := getEnv()
		log.WithFields(log.Fields{
			"envBootstrapConfigPath":   envBootstrapConfigPath,
			"envBootstrapConfigName":   envBootstrapConfigName,
			"envBootstrapConfigType":   envBootstrapConfigType,
			"envApplicationConfigPath": envApplicationConfigPath,
			"envApplicationConfigName": envApplicationConfigName,
			"envApplicationConfigType": envApplicationConfigType,
		}).Trace("从Env获取")
		if envBootstrapConfigPath != "" && envBootstrapConfigName != "" && envBootstrapConfigType != "" {
			//application文件位置
			log.Trace("从用户定义的位置读取application文件")
			applicationConfigPath, applicationConfigName, applicationConfigType =
				envApplicationConfigPath, envApplicationConfigName, envApplicationConfigType
			if err := newConfig(&root, envBootstrapConfigPath, envBootstrapConfigName, envBootstrapConfigType); err != nil {
				log.WithField("err", err).Error("func initBootstrap error")
				return Root{}, "", "", "", err
			}
		} else {
			// 从默认位置获取
			log.Trace("从默认位置读取application文件")
			if err := newConfig(&root, "", bootstrapName, configType); err != nil {
				log.WithField("err", err).Error("func initBootstrap error")
				return Root{}, "", "", "", err
			}
		}
	}
	log.WithFields(log.Fields{
		"root":                  root,
		"applicationConfigPath": applicationConfigPath,
		"applicationConfigName": applicationConfigName,
		"applicationConfigType": applicationConfigType,
	}).Trace("func initBootstrap end")
	return root, applicationConfigPath, applicationConfigName, applicationConfigType, nil
}

// 从命令行获取
func getFlag() (string, string, string, string, string, string, error) {
	log.Trace("func getFlag begin")
	pflag.String(FlagBootstrapConfigPath, "", "please input the bootstrap config file path")
	pflag.String(FlagBootstrapConfigName, bootstrapName, "please input the bootstrap config file name")
	pflag.String(FlagBootstrapConfigType, configType, "please input the bootstrap config file type")

	pflag.String(FlagApplicationConfigPath, "", "please input the application config file path")
	pflag.String(FlagApplicationConfigName, configName, "please input the application config file name")
	pflag.String(FlagApplicationConfigType, configType, "please input the application config file type")

	//获取标准包的flag
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	//BindFlag
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		log.WithField("err", err).Error("func getFlag error")
		return "", "", "", "", "", "", err
	}
	log.Trace("func getFlag end")
	return viper.GetString(FlagBootstrapConfigPath),
		viper.GetString(FlagBootstrapConfigName),
		viper.GetString(FlagBootstrapConfigType),
		viper.GetString(FlagApplicationConfigPath),
		viper.GetString(FlagApplicationConfigName),
		viper.GetString(FlagApplicationConfigType),
		nil
}

// 从ENV获取
func getEnv() (string, string, string, string, string, string) {
	log.Trace("func getEnv begin")
	v := viper.New()
	v.AutomaticEnv()
	log.Trace("func getEnv end")
	return v.GetString(EnvBootstrapConfigPath),
		v.GetString(EnvBootstrapConfigName),
		v.GetString(EnvBootstrapConfigType),
		v.GetString(EnvApplicationConfigPath),
		v.GetString(EnvApplicationConfigName),
		v.GetString(EnvApplicationConfigType)
}

// 读取本地文件
func newConfig(obj interface{}, path string, configName string, configType string) error {
	log.WithFields(log.Fields{
		"path":       path,
		"obj":        obj,
		"configName": configName,
		"configType": configType,
	}).Trace("func NewConfig begin")
	if path == "" {
		//获取项目的执行路径
		_path, err := os.Getwd()
		log.Trace("获取项目路径：%s", _path)
		if err != nil {
			log.WithField("err", err).Error("func NewConfig error")
			return err
		}
		path = _path
	}

	runtimeViper := viper.New()

	log.Trace("设置读取的文件路径")
	runtimeViper.AddConfigPath(path) //设置读取的文件路径

	log.Trace("设置读取的文件名")
	runtimeViper.SetConfigName(configName) //设置读取的文件名

	log.Trace("设置文件的类型")
	runtimeViper.SetConfigType(configType) //设置文件的类型

	//尝试进行配置读取
	log.Trace("尝试进行配置读取")
	if err := runtimeViper.ReadInConfig(); err != nil {
		log.WithField("err", err).Error("func NewConfig error")
		return err
	}

	log.Trace("Unmarshal")
	if err := runtimeViper.Unmarshal(obj); err != nil {
		log.WithField("err", err).Error("func NewConfig error")
		return err
	}
	log.WithField("obj", obj).Trace("func NewConfig end")
	return nil
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
