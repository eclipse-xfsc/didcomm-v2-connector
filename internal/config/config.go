package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

// constants for environment and mode
const (
	ENV_DEV  = "DEV"
	ENV_PROD = "PROD"

	HTTP   = "http"
	NATS   = "nats"
	HYBRID = "hybrid"
)

type TemplateConfiguration struct {
	Env             string `mapstructure:"env" envconfig:"DIDCOMMCONNECTOR_ENV" default:"DEV"`
	LogLevel        string `mapstructure:"logLevel" envconfig:"DIDCOMMCONNECTOR_LOGLEVEL" default:"info"`
	Port            int    `mapstructure:"port" envconfig:"DIDCOMMCONNECTOR_PORT"`
	Url             string `mapstructure:"url" envconfig:"DIDCOMMCONNECTOR_URL"`
	Label           string `mapstructure:"label" envconfig:"DIDCOMMCONNECTOR_LABEL"`
	TokenExpiration int    `mapstructure:"tokenExpiration" envconfig:"DIDCOMMCONNECTOR_TOKENEXPIRATION" default:"1"`
	DidComm         struct {
		ResolverUrl        string `mapstructure:"resolverUrl" envconfig:"DIDCOMMCONNECTOR_DIDCOMM_RESOLVERURL"`
		IsMessageEncrypted bool   `mapstructure:"messageEncrypted" envconfig:"DIDCOMMCONNECTOR_DIDCOMM_ISMESSAGEENCRYPTED"`
	} `mapstructure:"didcomm"`

	CloudForwarding struct {
		Protocol string `mapstructure:"protocol" envconfig:"DIDCOMMCONNECTOR_CLOUDFORWARDING_PROTOCOL" default:"nats"`
		Nats     struct {
			Url        string `mapstructure:"url" envconfig:"DIDCOMMCONNECTOR_CLOUDFORWARDING_NATS_URL"`
			Topic      string `mapstructure:"topic" envconfig:"DIDCOMMCONNECTOR_CLOUDFORWARDING_NATS_TOPIC"`
			QueueGroup string `mapstructure:"topic" envconfig:"DIDCOMMCONNECTOR_CLOUDFORWARDING_NATS_QUEUEGROUP"`
		} `mapstructure:"nats"`
		Http struct {
			Url string `mapstructure:"url" envconfig:"DIDCOMMCONNECTOR_CLOUDFORWARDING_HTTP_URL"`
		} `mapstructure:"http"`
	} `mapstructure:"messaging"`

	Database struct {
		InMemory bool   `mapstructure:"inMemory" envconfig:"DIDCOMMCONNECTOR_DATBASE_INMEMORY" default:"false"`
		Host     string `mapstructure:"host" envconfig:"DIDCOMMCONNECTOR_DATBASE_HOST"`
		Port     int    `mapstructure:"port" envconfig:"DIDCOMMCONNECTOR_DATBASE_PORT"`
		User     string `mapstructure:"user" envconfig:"DIDCOMMCONNECTOR_DATBASE_USER"`
		Password string `mapstructure:"password" envconfig:"DIDCOMMCONNECTOR_DATBASE_PASSWORD"`
		Keyspace string `mapstructure:"keyspace" envconfig:"DIDCOMMCONNECTOR_DATBASE_KEYSPACE"`
		DBName   string `mapstructure:"dbName" envconfig:"DIDCOMMCONNECTOR_DATBASE_DBNAME"`
	} `mapstructure:"db"`

	LoggerFile *os.File
}

var CurrentConfiguration TemplateConfiguration
var Logger *slog.Logger
var env string

func LoadConfig() error {
	slog.Info("Read Config")
	setDefaults()
	readConfig()
	if err := viper.Unmarshal(&CurrentConfiguration); err != nil {
		return err
	}
	slog.Info("Read ENV")
	err := envconfig.Process("DIDCOMMCONNECTOR", &CurrentConfiguration)

	if err != nil {
		Logger.Error("Error Processing ENVs", "config", err)
		return err
	}

	slog.Info("Set ENV")
	setEnvironment()
	if err := checkMode(); err != nil {
		return err
	}
	slog.Info("Set LogLevel")
	if err := setLogLevel(); err != nil {
		return err
	}
	slog.Info("Load Resolver")
	err = checkResolver(CurrentConfiguration.DidComm.ResolverUrl)
	if err != nil {
		Logger.Error("Resolver not available", "msg", err)
		return err
	} else {
		Logger.Info("Resolver available")
	}
	slog.Info("Marshal Config")
	b, err := json.Marshal(CurrentConfiguration)

	Logger.Info(string(b))

	if err != nil {
		Logger.Error("Config works not", "config", err)
		return err
	}

	return nil
}

func IsDev() bool {
	return env == ENV_DEV
}

func IsProd() bool {
	return env == ENV_PROD
}

func IsForwardTypeNats() bool {
	return CurrentConfiguration.CloudForwarding.Protocol == NATS
}

func IsForwardTypeHybrid() bool {
	return CurrentConfiguration.CloudForwarding.Protocol == HYBRID
}

func readConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("DIDCOMMCONNECTOR")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			slog.Warn("Configuration not found but environment variables will be taken into account.")
		}
	}
}

func setDefaults() {
	viper.SetDefault("env", ENV_DEV)
	viper.SetDefault("logLevel", "info")
	viper.SetDefault("port", 9090)
	viper.SetDefault("url", "http://localhost:9090")
	viper.SetDefault("cloudForwarding.type", "http")
	viper.SetDefault("didcomm.messageEncrypted", false)
}

func setEnvironment() {
	switch strings.ToUpper(CurrentConfiguration.Env) {
	case ENV_DEV:
		env = ENV_DEV
	case ENV_PROD:
		env = ENV_PROD
	default:
		env = ENV_DEV
	}
}

func setLogLevel() (err error) {
	logLevel := new(slog.LevelVar)
	switch strings.ToLower(CurrentConfiguration.LogLevel) {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "info":
		logLevel.Set(slog.LevelInfo)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		logLevel.Set(slog.LevelWarn)
	}

	if CurrentConfiguration.Env == ENV_PROD {
		// Create a folder named "logs" if it doesn't exist
		err := os.MkdirAll("logs", 0755)
		if err != nil {
			return err
		}
		fileName := filepath.Join("logs", "log_"+strconv.FormatInt(time.Now().Unix(), 10)+".log")
		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		CurrentConfiguration.LoggerFile = file

		Logger = slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{
			Level: logLevel,
		}))
		Logger.Info(fmt.Sprintf("log level set to %s", logLevel.Level().String()))
	} else {
		Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		}))
		Logger.Info(fmt.Sprintf("log level set to %s", logLevel.Level().String()))
	}

	return nil
}

func checkMode() error {
	selectedTyp := CurrentConfiguration.CloudForwarding.Protocol
	switch strings.ToLower(selectedTyp) {
	case HTTP:
	case NATS:
	case HYBRID:
		return fmt.Errorf("selected mode %s not yet supported", selectedTyp)
	default:
		return fmt.Errorf("unknown cloud forwarding type %s. Select one of these types: %s, %s or %s", selectedTyp, HTTP, NATS, HYBRID)
	}

	return nil
}

func checkResolver(resolverUrl string) error {
	queryUrl, err := url.JoinPath(resolverUrl, "/1.0/testIdentifiers")
	if err != nil {
		return err
	}
	resp, err := http.Get(queryUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("resolver not available")
	}
	return nil
}
