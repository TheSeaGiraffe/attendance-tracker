package config

import "github.com/TheSeaGiraffe/attendance-tracker/database"

type AppConfig struct {
	DBConfig database.DBConfig
}
