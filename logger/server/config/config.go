package config

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceID         string `mapstructure:"SERVICEID"`
	MongodbAddress    string `mapstructure:"MONGODB_ADDRESS"`
	RedisQueueAddress string `mapstructure:"REDIS_Q_ADDRESS"`
	MqExchangeKey     string `mapstructure:"MQ_EXCHANGE"`
	MqBacktestingKey  string `mapstructure:"MQ_BACKTESTING_KEY"`
	MqPreprocessKey   string `mapstructure:"MQ_PREPROCESS_KEY"`
	LocalLogFileDir   string `mapstructure:"LOCAL_LOG_FILE_DIR"`
	MqHost            string `mapstructure:"MQ_HOST"`
	MqPort            string `mapstructure:"MQ_PORT"`
	MqVHost           string `mapstructure:"MQ_VHOST"`
	MqUser            string `mapstructure:"MQ_USER"`
	MqPassword        string `mapstructure:"MQ_PASSWORD"`
}

var config_siongleton *ConfigSingleTon
var muonce sync.Once

type ConfigSingleTon struct {
	Config *Config
	mu     sync.RWMutex
}

func GetConfig() *Config {
	initConfig()
	config_siongleton.mu.RLock()
	defer config_siongleton.mu.RUnlock()
	return config_siongleton.Config
}

func initConfig() {
	if config_siongleton == nil {
		muonce.Do(func() {
			config_siongleton = &ConfigSingleTon{}
			if cf, err := loadConfig(); err == nil {
				config_siongleton.Config = cf
			} else {
				log.Fatal("error read app config file")
			}
			viper.WatchConfig()
			viper.OnConfigChange(func(e fsnotify.Event) {
				if cf, err := loadConfig(); err == nil {
					config_siongleton.Config = cf
				} else {
					log.Panic("failed to reload config file")
				}
			})
		})
	}
}

/*
單純回傳錯誤  由外部決定要不要Fatal, 畢竟有可能有替代方案
*/
func loadConfig() (cf *Config, err error) {
	config_siongleton.mu.Lock()
	defer config_siongleton.mu.Unlock()

	cf = &Config{}
	viper.SetConfigFile(fmt.Sprintf("%s/server/.env", getProjectRoot("github.com/RoyceAzure/rj/logger")))
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(cf)
	if err != nil {
		return
	}
	return
}

func getProjectRoot(moduleName string) string {
	// 執行 go list，但是加上額外的過濾條件
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", moduleName)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
