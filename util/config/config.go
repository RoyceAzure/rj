package config

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	viper "github.com/spf13/viper"
)

/*
把init log跟read log分開
init : 需要設置viper watch 與 onConfigChange
read log : 一般讀寫  需要使用讀寫所
*/
var config_siongleton *ConfigSingleTon
var muonce sync.Once

type ConfigSingleTon struct {
	Config *Config
	mu     sync.RWMutex
}

type Config struct {
	ModulerName               string `mapstructure:"MODULER_NAME"`
	DbConnectStr              string `mapstructure:"DB_CONECTION_STR"`
	MqExchangeKey             string `mapstructure:"MQ_EXCHANGE"`
	MqExchangeLoggerKey       string `mapstructure:"MQ_EXCHANGE_LOGGER"`
	MqBacktestingKey          string `mapstructure:"MQ_BACKTESTING_KEY"`
	MqPreprocessKey           string `mapstructure:"MQ_PREPROCESS_KEY"`
	MqLoggerKey               string `mapstructure:"MQ_LOGGER_KEY"`
	MinioBacktestBucket       string `mapstructure:"MINIO_BACK_TEST_BUCKET"`
	MinioBacktestMetaFileName string `mapstructure:"MINIO_BACK_TEST_META_FILE_NAME"`
	MinioPreprocessBucket     string `mapstructure:"MINIO_PRE_PROCESSED_BUCKET"`
	MinioHost                 string `mapstructure:"MINIO_HOST"`
	MinioPort                 string `mapstructure:"MINIO_PORT"`
	MinioUser                 string `mapstructure:"MINIO_ROOT_USER"`
	MinioPassword             string `mapstructure:"MINIO_ROOT_PASSWORD"`
	MqHost                    string `mapstructure:"MQ_HOST"`
	MqPort                    string `mapstructure:"MQ_PORT"`
	MqVHost                   string `mapstructure:"MQ_VHOST"`
	MqUser                    string `mapstructure:"MQ_USER"`
	MqPassword                string `mapstructure:"MQ_PASSWORD"`
	MqFullBacktestKEY         string `mapstructure:"MQ_FULL_BACKTEST_KEY"`
	LoggerSavePath            string `mapstructure:"LOGGER_SAVE_PATH"`
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
				log.Fatal("error read logger config")
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
	viper.SetConfigFile(fmt.Sprintf("%s/.env", getProjectRoot("github.com/RoyceAzure/rj")))
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
