package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
	"github.com/New-JAMneration/JAM-Protocol/logger"
	"io"
	"os"
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
		MaxTicketsPerBlock      int `json:"max_tickets_per_block"`
		TicketsPerValidator     int `json:"tickets_per_validator"`
		MaxBlocksHistory        int `json:"max_blocks_history"`
		AuthPoolMaxSize         int `json:"auth_pool_max_size"`
		AuthQueueSize           int `json:"auth_queue_size"`
		ValidatorsSuperMajority int `json:"validators_super_majority"`
		AvailBitfieldBytes      int `json:"avail_bitfield_bytes"`
	} `json:"const"`
}

func InitConfig() {
	var configPath string
	set := flag.NewFlagSet("config", flag.ExitOnError)
	set.StringVar(&configPath, "config", "example.json", "set config file path")
	err := set.Parse(os.Args[1:])
	if err != nil {
		return
	}

	if err := loadConfig(configPath); err != nil {
		panic(err)
	}
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

	initJamConst()
	initLog()
	initJamScaleRegistry()

	return nil
}

func initJamConst() {
	jam_types.ValidatorsCount = Config.Const.ValidatorsCount
	jam_types.CoresCount = Config.Const.CoresCount
	jam_types.EpochLength = Config.Const.EpochLength
	jam_types.MaxTicketsPerBlock = Config.Const.MaxTicketsPerBlock
	jam_types.TicketsPerValidator = Config.Const.TicketsPerValidator
	jam_types.MaxBlocksHistory = Config.Const.MaxBlocksHistory
	jam_types.AuthPoolMaxSize = Config.Const.AuthPoolMaxSize
	jam_types.AuthQueueSize = Config.Const.AuthQueueSize
	jam_types.ValidatorsSuperMajority = Config.Const.ValidatorsSuperMajority
	jam_types.AvailBitfieldBytes = Config.Const.AvailBitfieldBytes
}

func initLog() {
	logger.SetLevel(Config.Log.Level)
}

func initJamScaleRegistry() {
	jam_types.InitScaleRegistry()
}
