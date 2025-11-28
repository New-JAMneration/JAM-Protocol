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
		Level string `json:"level"`
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

func InitConfig(configPath string, mode string) error {
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

func loadConfig(path string) error {
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
	logger.SetLevel(Config.Log.Level)
}

func initJamScaleRegistry() {
	types.InitScaleRegistry()
}
