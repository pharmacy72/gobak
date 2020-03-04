package config

import (
	"github.com/pharmacy72/gobak/fileutils"
	"github.com/pharmacy72/gobak/level"
	"time"
)

type redisconfig struct {
	Enabled          bool   `json:"enabled"`
	Addr             string `json:"address"`
	Password         string `json:"password"`
	DB               int    `json:"db"`
	Timeout          int    `json:"timeout"` // default 1 sec
	Queue            int    `json:"queue"`   //default 100
	SendStatsEnabled bool
	PeriodStats      time.Duration `json:"periodstats"`
}

type Config struct {
	fileutils          *fileutils.FileUtils
	PathToNbackup      string       `json:"PathToNbackup"`
	PathToBackupFolder string       `json:"PathToBackupFolder"`
	DirectIO           bool         `json:"DirectIO"`
	AliasDb            string       `json:"AliasDb"`
	Password           string       `json:"Password"`
	User               string       `json:"User"`
	EmailFrom          string       `json:"EmailFrom"`
	EmailTo            string       `json:"EmailTo"`
	SMTPServer         string       `json:"SmtpServer"`
	Pathtogfix         string       `json:"Pathtogfix"`
	Physicalpathdb     string       `json:"Physicalpathdb"`
	NameBase           string       `json:"NameBase"`
	DeleteInt          int          `json:"DeleteInt"`
	TimeMsec           int          `json:"TimeMlsc"`
	Redis              *redisconfig `json:"redis"`
	Levels             []struct {
		Level int    `json:"level"`
		Tick  string `json:"tick"`
		Check bool   `json:"check"`
	} `json:"levels"`

	file         string
	LevelsConfig *level.Levels
}
