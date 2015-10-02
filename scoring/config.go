package scoring 

import (
	"encoding/json"
	"os"
)

type Config struct {
	RedisHost, PostgresHost string
	RedisPort, PostgresPort int
	PostgresDb, PostgresUser, PostgresPassword string
	TruthRoot, SubmissionRoot string
}

func LoadConfig(fn string) (Config, error) {
	file, err := os.Open(fn)
	if err!=nil {
		return Config{}, err
	}
	decoder := json.NewDecoder(file)
	var config Config
	err = decoder.Decode(&config)
	return config, err
}