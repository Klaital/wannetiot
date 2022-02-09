package config

import (
	"fmt"
	"github.com/MichaelS11/go-dht"
	"github.com/caarlos0/env/v6"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/klaital/max31855"
	"github.com/ryszard/sds011/go/sds011"
	log "github.com/sirupsen/logrus"
	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/spi/mcp3w0c"
	"periph.io/x/conn/v3/driver/driverreg"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
	"strings"
	"time"
)

type Config struct {
	// Metadata
	NodeName       string `env:"NODE_NAME" envDefault:"bedroom"`
	LogLevelStr    string `env:"LOG_LEVEL" envDefault:"debug"`
	LogLevel       log.Level
	Logger         *log.Logger
	WakeupDuration time.Duration `env:"WAKEUP_DURATION" envDefault:"30m"`

	// InfluxDB
	InfluxHost       string `env:"INFLUX_HOST"`
	InfluxToken      string `env:"INFLUX_TOKEN"`
	InfluxOrg        string `env:"INFLUX_ORG"`
	InfluxBucket     string `env:"INFLUX_BUCKET"`
	InfluxBufferSize int    `env:"INFLUX_BUFFER_LEN"`
	influxClient     api.WriteAPIBlocking

	// RF Remote Control
	RadioWaitTimeout   time.Duration `env:"RF_WAIT_TIMEOUT" envDefault:"1s"`
	RadioLatchResetPin string        `env:"RF_LATCH_RESET_PIN" envDefault:"GPIO17"`
	RadioChannelAPin   string        `env:"RF_CHANNEL_A_PIN" envDefault:"GPIO23"`
	RadioChannelBPin   string        `env:"RF_CHANNEL_B_PIN" envDefault:"GPIO25"`
	RadioChannelCPin   string        `env:"RF_CHANNEL_C_PIN" envDefault:"GPIO8"`
	RadioChannelDPin   string        `env:"RF_CHANNEL_D_PIN" envDefault:"GPIO24"`

	// Sensors
	PollInterval  time.Duration `env:"POLL_INTERVAL" envDefault:"5s"`
	AM2302Enabled bool          `env:"AM2302_ENABLED" envDefault:"false"`
	AM2302PinName string        `env:"AM2302" envDefault:"GPIO16"`
	AM2302Sensor  *dht.DHT
	SDS011Enabled bool   `env:"SDS011_ENABLED" envDefault:"false"`
	SDS011Txd     string `env:"SDS011_TXD" envDefault:"GPIO14"`
	SDS011Rxd     string `env:"SDS011_RXD" envDefault:"GPIO15"`
	SDSSerialPath string `env:"SDS_SERIAL_PATH" envDefault:"'/dev/ttyAMA0"`
	SDSSensor     *sds011.Sensor

	// Thermocouple
	Thermocouple1Enabled bool   `env:"THERMO1_ENABLED" envDefault:"false"`
	Thermocouple1Bus     string `env:"THERMO1_BUS" envDefault:"/dev/spidev0.1"`
	Thermocouple1CLKPin  string `env:"THERMO1_CLK" envDefault:"GPIO0"`
	Thermocouple1CSPin   string `env:"THERMO1_CS" envDefault:"GPIO5"`
	Thermocouple1MISOPin string `env:"THERMO1_MISO" envDefault:"GPIO0"`
	Thermocouple1        *max31855.Dev
	thermocouple1Spi     spi.PortCloser

	// Water Sensors
	WaterSensor1Enabled bool   `env:"WATER1_ENABLED" envDefault:"false"`
	WaterSensor1Pin     string `env:"WATER1_PIN" envDefault:"GPIO0"`
	WaterSensor1        gpio.PinIn
	WaterSensor2Enabled bool   `env:"WATER2_ENABLED" envDefault:"false"`
	WaterSensor2Pin     string `env:"WATER2_PIN" envDefault:"GPIO5"`
	WaterSensor2        gpio.PinIn

	// LED Light Strip
	LedStripEnabled    bool   `env:"LED_STRIP_ENABLED" envDefault:"true"`
	LedControlRed      string `env:"LED_CTRL_RED" envDefault:"GPIO2"`
	LedControlRedPin   gpio.PinOut
	LedControlGreen    string `env:"LED_CTRL_GREEN" envDefault:"GPIO3"`
	LedControlGreenPin gpio.PinOut
	LedControlWhite    string `env:"LED_CTRL_WHITE" envDefault:"GPIO4"`
	LedControlWhitePin gpio.PinOut
	LedControlBlue     string `env:"LED_CTRL_BLUE" envDefault:"GPIO18"`
	LedControlBluePin  gpio.PinOut

	// Control Panels
	ControlPanelsAdcClk int `env:"ADC1_CLK" envDefault:"5"` // Pin 29 / GPIO5
	//ControlPanelsAdcClkPin  gpio.PinIO
	ControlPanelsAdcCsz int `env:"ADC1_CSZ" envDefault:"21"` // Pin 40 / GPIO21
	//ControlPanelsAdcCszPin  gpio.PinIO
	ControlPanelsAdcDin int `env:"ADC1_DI" envDefault:"20"` // Pin 38 / GPIO20
	//ControlPanelsAdcDinPin  gpio.PinIO
	ControlPanelsAdcDout int `env:"ADC1_DO" envDefault:"19"` // Pin 35 / GPIO19
	//ControlPanelsAdcDoutPin gpio.PinIO
	GPIOChip         *gpiod.Chip
	ControlPanelsAdc *mcp3w0c.MCP3w0c

	Panel1Enabled        bool   `env:"PANEL1_ENABLED" envDefault:"false"`
	Panel1DimmerChannel  int    `env:"PANEL1_DIMMER_CHANNEL" envDefault:"0"`
	Panel1Pager          string `env:"PANEL1_PAGER" envDefault:"GPIO8"`
	Panel1PagerPin       gpio.PinIn
	Panel1LightSwitch    string `env:"PANEL1_LIGHTS" envDefault:"GPIO25"`
	Panel1LightSwitchPin gpio.PinIn
	Panel1LED            string `env:"PANEL1_LED" envDefault:"GPIO26"`
	Panel1LEDPin         gpio.PinOut
	Panel1Speaker        string `env:"PANEL1_SPEAKER" envDefault:"GPIO12"`
	Panel1SpeakerPin     gpio.PinOut
	Panel1Reset          string `env:"PANEL1_RESET" envDefault:"GPIO7"`
	Panel1ResetPin       gpio.PinOut

	Panel2Enabled        bool   `env:"PANEL2_ENABLED" envDefault:"false"`
	Panel2DimmerChannel  int    `env:"PANEL2_DIMMER_CHANNEL" envDefault:"1"`
	Panel2Pager          string `env:"PANEL2_PAGER" envDefault:"GPIO9"`
	Panel2PagerPin       gpio.PinIn
	Panel2LightSwitch    string `env:"PANEL2_LIGHTS" envDefault:"GPIO11"`
	Panel2LightSwitchPin gpio.PinIn
	Panel2LED            string `env:"PANEL2_LED" envDefault:"GPIO6"`
	Panel2LEDPin         gpio.PinOut
	Panel2Speaker        string `env:"PANEL2_SPEAKER" envDefault:"GPIO13"`
	Panel2SpeakerPin     gpio.PinOut
	Panel2Reset          string `env:"PANEL2_RESET" envDefault:"GPIO10"`
	Panel2ResetPin       gpio.PinOut
}

func (cfg *Config) GetInfluxDB() api.WriteAPIBlocking {
	if cfg.influxClient == nil {
		influx := influxdb2.NewClient(cfg.InfluxHost, cfg.InfluxToken)
		cfg.influxClient = influx.WriteAPIBlocking(cfg.InfluxOrg, cfg.InfluxBucket)
	}
	return cfg.influxClient
}

func (cfg *Config) InitSensors() {
	var err error
	// Temperature/Humidity
	if cfg.AM2302Enabled {
		if err = dht.HostInit(); err != nil {
			log.WithError(err).Fatal("Failed to initialize AM2302 temp/humidity sensor host")
		}
		cfg.AM2302Sensor, err = dht.NewDHT(cfg.AM2302PinName, dht.Fahrenheit, "")
		if err != nil {
			log.WithError(err).Fatal("Failed to initialize AM2302 temp/humidity sensor")
		}

		h, t, err := cfg.AM2302Sensor.Read()
		if err != nil {
			log.WithError(err).Fatal("Failed to take initial reading from AM2302 temp/humidity sensor")
		} else {
			log.WithFields(log.Fields{
				"temperature": t,
				"humidity":    h,
			}).Info("Initial temperature and humidity reading")
		}
	}

	// TODO: Air Quality
	if cfg.SDS011Enabled {
		cfg.SDSSensor, err = sds011.New(cfg.SDSSerialPath)
		if err != nil {
			cfg.Logger.WithError(err).Fatal("Failed to initialize dust sensor")
		}
		// Initial reading
		err = cfg.SDSSensor.Awake()
		if err != nil {
			cfg.Logger.WithError(err).Fatal("Failed to awaken dust sensor")
		}
		cfg.Logger.Debug("Waiting for sds011 sensor to wake up for initial reading")
		time.Sleep(2 * time.Second)
		p, err := cfg.SDSSensor.Get()
		if err != nil {
			cfg.Logger.WithError(err).Fatal("Failed to read dust sensor")
		}
		cfg.Logger.WithFields(log.Fields{
			"pm2.5": p.PM25,
			"pm1.0": p.PM10,
			"time":  p.Timestamp,
			"str":   p.String(),
		}).Info("Initial Dust reading complete")

		err = cfg.SDSSensor.Sleep()
		if err != nil {
			cfg.Logger.WithError(err).Fatal("Failed to sleep dust sensor")
		}

	}
}

func (cfg *Config) InitWaterSensors() error {
	if cfg.WaterSensor1Enabled {
		cfg.WaterSensor1 = gpioreg.ByName(cfg.WaterSensor1Pin)
		if cfg.WaterSensor1 == nil {
			log.WithField("pin", cfg.WaterSensor1Pin).Fatal("Failed to initialize WaterSensor1 pin")
		}
	}
	if cfg.WaterSensor2Enabled {
		cfg.WaterSensor2 = gpioreg.ByName(cfg.WaterSensor2Pin)
		if cfg.WaterSensor2 == nil {
			log.WithField("pin", cfg.WaterSensor2Pin).Fatal("Failed to initialize WaterSensor2 pin")
		}
	}
	return nil
}

func (cfg *Config) HaltWaterSensors() {
	if cfg.WaterSensor1 != nil {
		cfg.WaterSensor1.Halt()
	}
	if cfg.WaterSensor2 != nil {
		cfg.WaterSensor2.Halt()
	}
}

func spiDeviceList() string {
	spiList := make([]string, 0)
	for _, s := range spireg.All() {
		spiList = append(spiList, s.Name)
	}
	return fmt.Sprintf("[%s]", strings.Join(spiList, ", "))
}

func (cfg *Config) InitThermocouples() error {
	var err error
	logger := cfg.Logger.WithFields(log.Fields{
		"op":          "Config#InitThermocouples",
		"isEnabled":   cfg.Thermocouple1Enabled,
		"selectedBus": cfg.Thermocouple1Bus,
	})
	logger.Debug("Initializing thermocouples")
	if cfg.Thermocouple1Enabled {
		cfg.thermocouple1Spi, err = spireg.Open(cfg.Thermocouple1Bus)
		if err != nil {
			logger.WithField("availableSPI", spiDeviceList()).WithError(err).Error("Failed to open SPI bus")
			return err
		}
		cfg.Thermocouple1, err = max31855.New(cfg.thermocouple1Spi)
		if err != nil {
			logger.WithError(err).Error("Failed to initialize thermocouple device")
			return err
		}
	}
	return nil
}

func (cfg *Config) HaltThermocouples() {
	if cfg.Thermocouple1Enabled {
		cfg.thermocouple1Spi.Close()
	}
}

var cfg *Config

func New() *Config {
	if cfg == nil {
		cfg = new(Config)
		err := env.Parse(cfg)
		if err != nil {
			log.WithError(err).Fatal("Error loading env configs")
		}

		cfg.LogLevel, err = log.ParseLevel(cfg.LogLevelStr)
		if err != nil {
			log.WithError(err).WithField("level", cfg.LogLevelStr).Error("Failed to parse log level. Defaulting to INFO")
			cfg.LogLevel = log.InfoLevel
		}
		log.WithField("LogLevel", cfg.LogLevel.String()).Info("Setting log level")
		log.SetLevel(cfg.LogLevel)
		cfg.Logger = log.New()
		cfg.Logger.SetLevel(cfg.LogLevel)
	}

	return cfg
}

func (cfg *Config) InitPins() {
	var err error

	if cfg.Panel1Enabled || cfg.Panel2Enabled {
		log.Debug("initializing ADC")
		cfg.GPIOChip, err = gpiod.NewChip("gpiochip0", gpiod.WithConsumer("mbed"))
		if err != nil {
			log.WithField("chips", gpiod.Chips()).WithError(err).Fatal("failed to open gpio chip")
		}
		cfg.ControlPanelsAdc, err = mcp3w0c.NewMCP3008(cfg.GPIOChip, cfg.ControlPanelsAdcClk, cfg.ControlPanelsAdcCsz, cfg.ControlPanelsAdcDin, cfg.ControlPanelsAdcDout)
		if err != nil {
			log.WithError(err).Fatal("Failed to initialize ADC")
		}
	}

	_, err = host.Init()
	if err != nil {
		log.WithError(err).Fatal("Failed to init host")
	}

	if _, err = driverreg.Init(); err != nil {
		log.WithError(err).Fatal("Failed to init driverreg")
	}

	if cfg.LedStripEnabled {
		cfg.LedControlRedPin = gpioreg.ByName(cfg.LedControlRed)
		if cfg.LedControlRedPin == nil {
			log.WithField("pin", cfg.LedControlRedPin.String()).Error("Failed to init Red LED pin")
		}
		if err = cfg.LedControlRedPin.Out(gpio.Low); err != nil {
			log.WithField("pin", cfg.LedControlRedPin.String()).WithError(err).Error("Error with initial Red LED settings")
		}
		cfg.LedControlGreenPin = gpioreg.ByName(cfg.LedControlGreen)
		if cfg.LedControlGreenPin == nil {
			log.WithField("pin", cfg.LedControlGreenPin.String()).Error("Failed to init Green LED pin")
		}
		if err = cfg.LedControlGreenPin.Out(gpio.Low); err != nil {
			log.WithField("pin", cfg.LedControlGreenPin.String()).WithError(err).Error("Error with initial Green LED settings")
		}
		cfg.LedControlWhitePin = gpioreg.ByName(cfg.LedControlWhite)
		if cfg.LedControlWhitePin == nil {
			log.WithField("pin", cfg.LedControlWhitePin.String()).Error("Failed to init White LED pin")
		}
		if err = cfg.LedControlWhitePin.Out(gpio.Low); err != nil {
			log.WithField("pin", cfg.LedControlWhitePin.String()).WithError(err).Error("Error with initial White LED settings")
		}
		cfg.LedControlBluePin = gpioreg.ByName(cfg.LedControlBlue)
		if cfg.LedControlBluePin == nil {
			log.WithField("pin", cfg.LedControlBluePin.String()).Fatal("Failed to init Blue LED pin")
		}
		if err = cfg.LedControlBluePin.Out(gpio.Low); err != nil {
			log.WithField("pin", cfg.LedControlBluePin.String()).WithError(err).Error("Error with initial Blue LED settings")
		}
	}

	if cfg.Panel1Enabled {
		// Configure the Control Panel direct IO pins
		cfg.Panel1PagerPin = gpioreg.ByName(cfg.Panel1Pager)
		if cfg.Panel1PagerPin == nil {
			log.Fatal("Failed to init Panel1PagerPin")
		}
		if err = cfg.Panel1PagerPin.In(gpio.PullDown, gpio.RisingEdge); err != nil {
			log.WithError(err).Fatal("Failed to configure Panel1PagerPin")
		}
		cfg.Panel1LightSwitchPin = gpioreg.ByName(cfg.Panel1LightSwitch)
		if cfg.Panel1LightSwitchPin == nil {
			log.Fatal("Failed to init Panel1LightSwitchPin")
		}
		if err = cfg.Panel1LightSwitchPin.In(gpio.PullDown, gpio.RisingEdge); err != nil {
			log.WithError(err).Fatal("Failed to configure Panel1LightSwitchPin")
		}
		cfg.Panel1LEDPin = gpioreg.ByName(cfg.Panel1LED)
		if cfg.Panel1LEDPin == nil {
			log.Fatal("Failed to init Panel1LEDPin")
		}
		cfg.Panel1LEDPin.Out(gpio.Low)
		cfg.Panel1SpeakerPin = gpioreg.ByName(cfg.Panel1Speaker)
		if cfg.Panel1SpeakerPin == nil {
			log.Fatal("Failed to init Panel1SpeakerPin")
		}
		cfg.Panel1SpeakerPin.Out(gpio.Low)
		cfg.Panel1ResetPin = gpioreg.ByName(cfg.Panel1Reset)
		if cfg.Panel1ResetPin == nil {
			log.Fatal("Failed to init Panel1ResetPin")
		}
		cfg.Panel1ResetPin.Out(gpio.Low)
	}

	if cfg.Panel2Enabled {
		cfg.Panel2PagerPin = gpioreg.ByName(cfg.Panel2Pager)
		if cfg.Panel2PagerPin == nil {
			log.Fatal("Failed to init Panel2PagerPin")
		}
		if err = cfg.Panel2PagerPin.In(gpio.PullDown, gpio.RisingEdge); err != nil {
			log.WithError(err).Fatal("Failed to configure Panel2PagerPin")
		}
		cfg.Panel2LightSwitchPin = gpioreg.ByName(cfg.Panel2LightSwitch)
		if cfg.Panel2LightSwitchPin == nil {
			log.Fatal("Failed to init Panel2LightSwitchPin")
		}
		if err = cfg.Panel2LightSwitchPin.In(gpio.PullDown, gpio.RisingEdge); err != nil {
			log.WithError(err).Fatal("Failed to configure Panel2LightSwitchPin")
		}
		cfg.Panel2LEDPin = gpioreg.ByName(cfg.Panel2LED)
		if cfg.Panel2LEDPin == nil {
			log.Fatal("Failed to init Panel2LEDPin")
		}
		cfg.Panel2LEDPin.Out(gpio.Low)
		cfg.Panel2SpeakerPin = gpioreg.ByName(cfg.Panel2Speaker)
		if cfg.Panel2SpeakerPin == nil {
			log.Fatal("Failed to init Panel2SpeakerPin")
		}
		cfg.Panel2SpeakerPin.Out(gpio.Low)
		cfg.Panel2ResetPin = gpioreg.ByName(cfg.Panel2Reset)
		if cfg.Panel2ResetPin == nil {
			log.Fatal("Failed to init Panel2ResetPin")
		}
		cfg.Panel2ResetPin.Out(gpio.Low)
	}
}

// HaltPins will call pin.Halt on all pins used for PWM output.
// This function must be called before the program exits, or else
// the Raspberry Pi must be reset before PWM can be used again.
func (cfg *Config) HaltPins() {
	if cfg.LedControlRedPin != nil {
		cfg.LedControlRedPin.Halt()
	}
	if cfg.LedControlGreenPin != nil {
		cfg.LedControlGreenPin.Halt()
	}
	if cfg.LedControlWhitePin != nil {
		cfg.LedControlWhitePin.Halt()
	}
	if cfg.LedControlBluePin != nil {
		cfg.LedControlBluePin.Halt()
	}
	if cfg.Panel1SpeakerPin != nil {
		cfg.Panel1SpeakerPin.Halt()
	}
	if cfg.Panel2SpeakerPin != nil {
		cfg.Panel2SpeakerPin.Halt()
	}
	if cfg.GPIOChip != nil {
		cfg.GPIOChip.Close()
	}
	if cfg.ControlPanelsAdc != nil {
		cfg.ControlPanelsAdc.Close()
	}
}
