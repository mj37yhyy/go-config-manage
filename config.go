package go_config_manage

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"os"
	"time"
)

const (
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

type Remote struct {
	Enabled       bool     `mapstructure:"enabled"`
	Provider      string   `mapstructure:"provider"`
	Endpoint      string   `mapstructure:"endpoint"`
	Path          []string `mapstructure:"path"`
	SecretKeyring string   `mapstructure:"secret-keyring"`
}

type Config struct {
	Remote Remote `mapstructure:"remote"`
	File   bool   `mapstructure:"file"`
}

func InitConfig(obj interface{}) error {
	var root = Root{}
	if err := NewConfig(&root, bootstrapName, configType); err != nil {
		return err
	}
	if root.Application.Config.File {
		if err := NewConfig(&obj, configName, configType); err != nil {
			return err
		}
	}
	if root.Application.Config.Remote.Enabled {
		if root.Application.Config.Remote.SecretKeyring != "" {
			for _, path := range root.Application.Config.Remote.Path {
				if err := NewRemoteConfig(
					func(runtimeViper *viper.Viper) error {
						if err := runtimeViper.AddSecureRemoteProvider(
							root.Application.Config.Remote.Provider,
							root.Application.Config.Remote.Endpoint,
							path,
							root.Application.Config.Remote.SecretKeyring,
						); err != nil {
							return err
						}

						return nil
					},
					&obj,
				); err != nil {
					return err
				}
			}
		} else {
			for _, path := range root.Application.Config.Remote.Path {
				if err := NewRemoteConfig(
					func(runtimeViper *viper.Viper) error {
						if err := runtimeViper.AddRemoteProvider(
							root.Application.Config.Remote.Provider,
							root.Application.Config.Remote.Endpoint,
							path,
						); err != nil {
							return err
						}
						return nil
					},
					&obj,
				); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func NewRemoteConfig(addProvider func(runtimeViper *viper.Viper) error, obj interface{}) error {
	runtimeViper := viper.New()
	if err := addProvider(runtimeViper); err != nil {
		return err
	}
	runtimeViper.SetConfigType(configType)

	// 第一次从远程配置中读取
	if err := runtimeViper.ReadRemoteConfig(); err != nil {
		return err
	}
	if err := runtimeViper.Unmarshal(&obj); err != nil {
		return err
	}

	go func() {
		for {
			time.Sleep(time.Second * 5) // 每次请求后延迟
			err := runtimeViper.WatchRemoteConfig()
			if err != nil {
				log.Errorf("unable to read remote config: %v", err)
				continue
			}

			//将新配置解组到我们的运行时配置结构中。您还可以使用通道
			//实现信号以通知系统更改
			if err := runtimeViper.Unmarshal(&obj); err != nil {
				log.Errorf("unmarshal config error: %v", err)
				continue
			}
		}
	}()
	return nil
}

func NewConfig(obj interface{}, configName string, configType string) error {
	//获取项目的执行路径
	path, err := os.Getwd()
	if err != nil {
		return err
	}

	runtimeViper := viper.New()

	runtimeViper.AddConfigPath(path)       //设置读取的文件路径
	runtimeViper.SetConfigName(configName) //设置读取的文件名
	runtimeViper.SetConfigType(configType) //设置文件的类型
	//尝试进行配置读取
	if err := runtimeViper.ReadInConfig(); err != nil {
		return err
	}
	if err := runtimeViper.Unmarshal(&obj); err != nil {
		return err
	}
	return nil
}
