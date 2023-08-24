package configs

import (
	"encoding/json"
	"main/pkg/models"
	"os"
)

func InitConfigs() (*models.Config, error) {
	bytes, err := os.ReadFile("./internal/configs/config.json")
	if err != nil {
		return nil, err
	}
	var config models.Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
