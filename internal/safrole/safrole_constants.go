package safrole

type Config struct {
	ValidatorsCount int
	EpochLength     int
}

var ConfigFull = Config{
	ValidatorsCount: 1023,
	EpochLength:     600,
}

var ConfigTiny = Config{
	ValidatorsCount: 6,
	EpochLength:     12,
}
