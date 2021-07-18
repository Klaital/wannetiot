package config

import (
	"github.com/Netflix/go-env"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type Config struct {
	// Metadata
	NodeName string `env:"NODE_NAME"`
	Location string `env:"LOCATION"`

	// Logging
	LogLevel   string `env:"LOG_LEVEL"`
	LogContext *log.Entry

	// Databases
	InfluxHost   string `env:"INFLUX_HOST"`
	InfluxToken  string `env:"INFLUX_TOKEN"`
	InfluxOrg    string `env:"INFLUX_ORG"`
	InfluxBucket string `env:"INFLUX_BUCKET"`
	influxClient influxdb2.Client

	// Sensors
	SampleIntervalStr string `env:"SAMPLE_INTERVAL"`
	SampleInterval    time.Duration
	ChipID            string `env:"CHIP_ID" envDefault:"gpiochip0"`
	TempSensorPin     int    `env:"TEMP_ADC_ID"`

	// TODO: The ADC config
}

// Load parses the environment variables to populate a Config struct.
// TODO: make this cache its results in a singleton
func Load() *Config {
	logger := log.StandardLogger()
	logger.SetLevel(log.DebugLevel)
	logger.SetFormatter(&log.JSONFormatter{
		PrettyPrint: true,
	})
	logger.SetReportCaller(true) // TODO: only set this in test and dev, not production
	envfile := os.Getenv("ENV_FILE")
	if len(envfile) == 0 {
		envfile = ".env"
	}
	err := godotenv.Load(envfile)
	if err != nil {
		logger.WithError(err).WithField("envfile", envfile).Fatal("Failed to load env file")
	}
	var cfg Config
	_, err = env.UnmarshalFromEnviron(&cfg)
	if err != nil {
		logger.Fatal(err)
	}

	// Configure the standard logger
	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err == nil {
		logger.SetLevel(logLevel)
	} else {
		logger.SetLevel(log.ErrorLevel)
		logger.WithError(err).Error("Failed to parse loglevel. Defaulting to Error level")
	}
	// Only set the calling function/method as a field in debug because
	// this comes with considerable overhead.
	logger.SetReportCaller(logLevel == log.DebugLevel)

	cfg.LogContext = logger.WithFields(log.Fields{
		"Location": cfg.Location,
		"Node":     cfg.NodeName,
	})

	// Connect to Influx DB and cache the connection pool
	cfg.ConnectInflux()

	// Parse the sample duration
	cfg.SampleInterval, err = time.ParseDuration(cfg.SampleIntervalStr)
	if err != nil {
		cfg.LogContext.WithError(err).WithField("SampleInterval", cfg.SampleIntervalStr).Fatal("Failed to parse Sample Interval")
	}

	return &cfg
}

func (cfg *Config) ConnectInflux() influxdb2.Client {
	if cfg.influxClient == nil {
		influx := influxdb2.NewClient(cfg.InfluxHost, cfg.InfluxToken)
		cfg.influxClient = influx
	}
	return cfg.influxClient
}
