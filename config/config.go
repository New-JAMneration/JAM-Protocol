package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

var Config config

type config struct {
	Log struct {
		Level      string `json:"level"`
		Color      bool   `json:"color"`
		Enabled    bool   `json:"enabled"`
		PVM        bool   `json:"pvm"`
		TimeFormat string `json:"time_format"` // Optional: time format for log timestamps
	} `json:"log"`
	Const struct {
		ValidatorsCount         int `json:"validators_count"`
		CoresCount              int `json:"cores_count"`
		EpochLength             int `json:"epoch_length"`
		SlotSubmissionEnd       int `json:"slot_submission_end"`
		MaxTicketsPerBlock      int `json:"max_tickets_per_block"`
		TicketsPerValidator     int `json:"tickets_per_validator"`
		ValidatorsSuperMajority int `json:"validators_super_majority"`
		AvailBitfieldBytes      int `json:"avail_bitfield_bytes"`
	} `json:"const"`
	Redis struct {
		Address  string `json:"address"`
		Port     int    `json:"port"`
		Password string `json:"password"`
	} `json:"redis"`
	Info struct {
		FuzzVersion  uint8  `json:"fuzz_version"`
		FuzzFeatures uint32 `json:"fuzz_features"`
		JamVersion   string `json:"jam_version"`
		AppVersion   string `json:"app_version"`
		Name         string `json:"name"`
	} `json:"info"`
	FolderWise bool
}

func DefaultConfig() config {
	return config{
		Log: struct {
			Level      string `json:"level"`
			Color      bool   `json:"color"`
			Enabled    bool   `json:"enabled"`
			PVM        bool   `json:"pvm"`
			TimeFormat string `json:"time_format"`
		}{
			Level:      "DEBUG",        // Default: show all logs
			Color:      true,           // Default: colored output
			Enabled:    true,           // Default: logging enabled
			PVM:        false,          // Default: PVM logging disabled
			TimeFormat: "15:04:05.000", // Default: HH:MM:SS.ms format
		},
		Const: struct {
			ValidatorsCount         int `json:"validators_count"`
			CoresCount              int `json:"cores_count"`
			EpochLength             int `json:"epoch_length"`
			SlotSubmissionEnd       int `json:"slot_submission_end"`
			MaxTicketsPerBlock      int `json:"max_tickets_per_block"`
			TicketsPerValidator     int `json:"tickets_per_validator"`
			ValidatorsSuperMajority int `json:"validators_super_majority"`
			AvailBitfieldBytes      int `json:"avail_bitfield_bytes"`
		}{
			ValidatorsCount:         6,
			CoresCount:              2,
			EpochLength:             12,
			SlotSubmissionEnd:       10,
			MaxTicketsPerBlock:      3,
			TicketsPerValidator:     3,
			ValidatorsSuperMajority: 5,
			AvailBitfieldBytes:      1,
		},
		Redis: struct {
			Address  string `json:"address"`
			Port     int    `json:"port"`
			Password string `json:"password"`
		}{
			Address:  "localhost",
			Port:     6379,
			Password: "password",
		},
		Info: struct {
			FuzzVersion  uint8  `json:"fuzz_version"`
			FuzzFeatures uint32 `json:"fuzz_features"`
			JamVersion   string `json:"jam_version"`
			AppVersion   string `json:"app_version"`
			Name         string `json:"name"`
		}{
			FuzzVersion:  1,
			FuzzFeatures: 2,
			JamVersion:   "0.7.1",
			AppVersion:   "0.1.0",
			Name:         "new_jamneration",
		},
	}
}

func InitConfig(configPath string, mode string) error {
	Config = DefaultConfig()
	if err := loadConfig(configPath); err != nil {
		panic(err)
	}

	switch mode {
	case "full":
		types.SetFullMode()
	case "tiny":
		types.SetTinyMode()
	case "custom":
		initJamConst()
	default:
		return fmt.Errorf("invalid mode: %s", mode)
	}
	return nil
}

func UpdateVersion(jamVersion, appVersion string) {
	Config.Info.JamVersion = jamVersion
	Config.Info.AppVersion = appVersion
}

func loadConfig(path string) error {
	// If the config file exists, load it
	if _, err := os.Stat(path); err == nil {
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("can't open config file: %w", err)
		}
		defer file.Close()

		bytes, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("can't read config file: %w", err)
		}

		if err := json.Unmarshal(bytes, &Config); err != nil {
			return fmt.Errorf("can't parse config file: %w", err)
		}
	} else {
		// config file does not exist, use default config
		logger.Warn("Config file not found, using default configuration")
	}

	initLog()
	initJamScaleRegistry()
	return nil
}

func initJamConst() {
	types.ValidatorsCount = Config.Const.ValidatorsCount
	types.CoresCount = Config.Const.CoresCount
	types.EpochLength = Config.Const.EpochLength
	types.SlotSubmissionEnd = Config.Const.SlotSubmissionEnd
	types.MaxTicketsPerBlock = Config.Const.MaxTicketsPerBlock
	types.TicketsPerValidator = Config.Const.TicketsPerValidator
	types.ValidatorsSuperMajority = Config.Const.ValidatorsSuperMajority
	types.AvailBitfieldBytes = Config.Const.AvailBitfieldBytes
}

func initLog() {
	// Configure main logger from config file
	logger.ConfigureLogger("main", logger.LoggerConfig{
		Level:      Config.Log.Level,
		Enabled:    Config.Log.Enabled,
		Color:      Config.Log.Color,
		TimeFormat: Config.Log.TimeFormat, // Optional: per-logger override
	})

	// Configure PVM logger from config file
	logger.ConfigureLogger("pvm", logger.LoggerConfig{
		Level:      Config.Log.Level, // Use same level as main
		Enabled:    Config.Log.PVM,
		Color:      Config.Log.Color,
		TimeFormat: Config.Log.TimeFormat, // Optional: per-logger override
	})
}

// InitLogConfig loads ONLY the `log` section from the given config file and
// configures the loggers. Other sections (const/redis/info/...) are ignored.
// If the file does not exist or `log` is missing, defaults from DefaultConfig
// are used.
func InitLogConfig(path string) {
	// Start from defaults
	cfg := DefaultConfig()

	// Try to read just the `log` section from the file (if it exists)
	if _, err := os.Stat(path); err == nil {
		file, err := os.Open(path)
		if err == nil {
			defer file.Close()

			bytes, err := io.ReadAll(file)
			if err == nil {
				// Local struct that only cares about `log`
				var partial struct {
					Log struct {
						Level      string `json:"level"`
						Color      bool   `json:"color"`
						Enabled    bool   `json:"enabled"`
						PVM        bool   `json:"pvm"`
						TimeFormat string `json:"time_format"`
					} `json:"log"`
				}

				if err := json.Unmarshal(bytes, &partial); err == nil {
					// Override defaults only for log section
					cfg.Log = partial.Log
				}
			}
		}
	}

	// Apply to global Config.Log so other code can read it if needed
	Config.Log = cfg.Log

	// Configure loggers based on cfg.Log
	logger.ConfigureLogger("main", logger.LoggerConfig{
		Level:      cfg.Log.Level,
		Enabled:    cfg.Log.Enabled,
		Color:      cfg.Log.Color,
		TimeFormat: cfg.Log.TimeFormat,
	})

	logger.ConfigureLogger("pvm", logger.LoggerConfig{
		Level:      cfg.Log.Level,
		Enabled:    cfg.Log.PVM,
		Color:      cfg.Log.Color,
		TimeFormat: cfg.Log.TimeFormat,
	})
}

func initJamScaleRegistry() {
	types.InitScaleRegistry()
}
