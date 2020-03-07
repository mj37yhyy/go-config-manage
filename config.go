package go_config_manage

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"os"
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
	Enabled       bool     `mapstructure:"enabled"`
	Provider      string   `mapstructure:"provider"`
	Endpoint      []string `mapstructure:"endpoint"`
	Path          []string `mapstructure:"path"`
	SecretKeyring string   `mapstructure:"secret-keyring"`
}

func InitConfig(obj interface{}) error {
	root, applicationConfigPath, applicationConfigName, applicationConfigType, err := initBootstrap()
	if err != nil {
		return err
	}

	err = initApplication(root, applicationConfigPath, applicationConfigName, applicationConfigType, &obj)
	if err != nil {
		return err
	}
	return nil
}

func initApplication(root Root, applicationConfigPath string, applicationConfigName string, applicationConfigType string, obj interface{}) error {
	if root.Application.Config.File {
		if applicationConfigPath != "" && applicationConfigName != "" && applicationConfigType != "" {
			// 如果传入的application位置存在
			if err := NewConfig(&obj, applicationConfigPath, applicationConfigName, applicationConfigType); err != nil {
				return err
			}
		} else if err := NewConfig(&obj, "", configName, configType); err != nil {
			return err
		}
	}

	// 从远程获取
	for _, remote := range root.Application.Config.Remote {
		if remote.Enabled {
			for _, path := range remote.Path {
				if remote.SecretKeyring != "" {
					err := AddSecureRemoteProvider(remote, path, obj)
					if err != nil {
						return err
					}
				} else {
					err := AddRemoteProvider(remote, path, obj)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func initBootstrap() (Root, string, string, string, error) {
	var root = Root{}
	var applicationConfigPath, applicationConfigName, applicationConfigType string
	// 从命令行获取
	flagBootstrapConfigPath, flagBootstrapConfigName, flagBootstrapConfigType,
		flagApplicationConfigPath, flagApplicationConfigName, flagApplicationConfigType,
		err := getFlag()
	if err != nil {
		return Root{}, "", "", "", err
	}
	if flagBootstrapConfigPath != "" && flagBootstrapConfigName != "" && flagBootstrapConfigType != "" {
		//application文件位置
		applicationConfigPath, applicationConfigName, applicationConfigType =
			flagApplicationConfigPath, flagApplicationConfigName, flagApplicationConfigType
		if err := NewConfig(&root, flagBootstrapConfigPath, flagBootstrapConfigName, flagBootstrapConfigType); err != nil {
			return Root{}, "", "", "", err
		}
	} else {
		// 从env获取
		envBootstrapConfigPath, envBootstrapConfigName, envBootstrapConfigType,
			envApplicationConfigPath, envApplicationConfigName, envApplicationConfigType := getEnv()
		if envBootstrapConfigPath != "" && envBootstrapConfigName != "" && envBootstrapConfigType != "" {
			//application文件位置
			applicationConfigPath, applicationConfigName, applicationConfigType =
				envApplicationConfigPath, envApplicationConfigName, envApplicationConfigType
			if err := NewConfig(&root, envBootstrapConfigPath, envBootstrapConfigName, envBootstrapConfigType); err != nil {
				return Root{}, "", "", "", err
			}
		} else {
			// 从默认位置获取
			if err := NewConfig(&root, "", bootstrapName, configType); err != nil {
				return Root{}, "", "", "", err
			}
		}
	}
	return root, applicationConfigPath, applicationConfigName, applicationConfigType, nil
}

func getFlag() (string, string, string, string, string, string, error) {
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
		return "", "", "", "", "", "", err
	}
	return viper.GetString(FlagBootstrapConfigPath),
		viper.GetString(FlagBootstrapConfigName),
		viper.GetString(FlagBootstrapConfigType),
		viper.GetString(FlagApplicationConfigPath),
		viper.GetString(FlagApplicationConfigName),
		viper.GetString(FlagApplicationConfigType),
		nil
}

func getEnv() (string, string, string, string, string, string) {
	v := viper.New()
	v.AutomaticEnv()
	return v.GetString(EnvBootstrapConfigPath),
		v.GetString(EnvBootstrapConfigName),
		v.GetString(EnvBootstrapConfigType),
		v.GetString(EnvApplicationConfigPath),
		v.GetString(EnvApplicationConfigName),
		v.GetString(EnvApplicationConfigType)
}

func AddRemoteProvider(remote Remote, path string, obj interface{}) error {
	if err := NewRemoteConfig(
		func(runtimeViper *viper.Viper) error {
			for _, endpoint := range remote.Endpoint {
				if err := runtimeViper.AddRemoteProvider(
					remote.Provider,
					endpoint,
					path,
				); err != nil {
					return err
				}
			}
			return nil
		},
		&obj,
	); err != nil {
		return err
	}
	return nil
}

func AddSecureRemoteProvider(remote Remote, path string, obj interface{}) error {
	if err := NewRemoteConfig(
		func(runtimeViper *viper.Viper) error {
			for _, endpoint := range remote.Endpoint {
				if err := runtimeViper.AddSecureRemoteProvider(
					remote.Provider,
					endpoint,
					path,
					remote.SecretKeyring,
				); err != nil {
					return err
				}
			}
			return nil
		},
		&obj,
	); err != nil {
		return err
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

func NewConfig(obj interface{}, path string, configName string, configType string) error {
	if path == "" {
		//获取项目的执行路径
		_path, err := os.Getwd()
		if err != nil {
			return err
		}
		path = _path
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
