package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pharmacy72/gobak/fileutils"
	"github.com/pharmacy72/gobak/level"
)
const EnvPrefix = "GOBAK"

//Errors in the configuration file
var (
	ErrFolderBackupNotExists = errors.New("Config: Folder for backup not found")
	ErrConfigLevel           = errors.New("Config: levels not found")
	ErrNbackupNotExists      = errors.New("Config: file Nbackup destination not exists")
	ErrGfixNotExists         = errors.New("Config: file gfix  destination not exists")
	ErrPhysicalNotExists     = errors.New("Config: Physicalpathdb destination not exists")
	ErrAliasDBNotExists      = errors.New("Config: Alias DB is empty")
)

var cfg *Config

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

//A Config it contains the application settings from a file config.json
type Config struct {
	PathToNbackup      string      `json:"PathToNbackup"`
	PathToBackupFolder string      `json:"PathToBackupFolder"`
	DirectIO           bool        `json:"DirectIO"`
	AliasDb            string      `json:"AliasDb"`
	Password           string      `json:"Password"`
	User               string      `json:"User"`
	EmailFrom          string      `json:"EmailFrom"`
	EmailTo            string      `json:"EmailTo"`
	SMTPServer         string      `json:"SmtpServer"`
	Pathtogfix         string      `json:"Pathtogfix"`
	Physicalpathdb     string      `json:"Physicalpathdb"`
	NameBase           string      `json:"NameBase"`
	DeleteInt          int         `json:"DeleteInt"`
	TimeMsec           int         `json:"TimeMlsc"`
	Redis              redisconfig `json:"redis"`
	Levels             []struct {
		Level int    `json:"level"`
		Tick  string `json:"tick"`
		Check bool   `json:"check"`
	} `json:"levels"`

	file         string
	LevelsConfig *level.Levels
}

//Current returns a *Config each time one and the same or or will be creating it
func Current() *Config {
	if cfg == nil {
		fileconfig := "config.json"

		if _, e := os.Stat(fileconfig); e != nil && os.IsNotExist(e) {
			fileconfig = filepath.Join(filepath.Dir(os.Args[0]), "config.json")
		}
		if cfg == nil {
			var err error
			cfg, err = loadConfig(fileconfig)
			if err != nil {
				panic(err)
			}
			cfg.file = fileconfig
		}
	}
	return cfg
}

func nvl(val, def interface{}) interface{} {
	if val == nil {
		return def
	}
	return val
}

func lookupEnv(key string) (result string) {
	result, _ = os.LookupEnv(key)
	return result
}

func loadConfig(filename string) (result *Config, err error) {
	file, e := ioutil.ReadFile(filename)
	if e != nil {
		return nil, e
	}
	result = &Config{}
	var res map[string]interface{}
	if e = json.Unmarshal(file, &res); e != nil {
		return nil, e
	}
	result.PathToNbackup = res["PathToNbackup"].(string)
	result.PathToBackupFolder = filepath.Clean(nvl(res["PathToBackupFolder"], "").(string))
	result.AliasDb = strings.TrimSpace(nvl(res["AliasDb"], "").(string))
	result.Physicalpathdb = nvl(res["Physicalpathdb"], "").(string)
	result.Password = nvl(res["Password"], lookupEnv("ISC_PASSWORD")).(string)
	result.User = nvl(res["User"], lookupEnv("ISC_USER")).(string)
	result.EmailFrom = nvl(res["EmailFrom"], "").(string)
	result.EmailTo = nvl(res["EmailTo"], "").(string)
	result.SMTPServer = nvl(res["SmtpServer"], "").(string)
	result.Pathtogfix = nvl(res["Pathtogfix"], "").(string)
	result.NameBase = nvl(res["NameBase"], "").(string)
	result.DeleteInt = int(nvl(res["DeleteInt"], 90).(float64))
	result.TimeMsec = int(nvl(res["TimeMlsc"], 10000).(float64))
	result.DirectIO = nvl(res["DirectIO"], false).(bool)
	result.LevelsConfig = level.NewList()

	for _, p := range res["levels"].([]interface{}) {
		litem := p.(map[string]interface{})
		cfg, err := result.LevelsConfig.Add(
			level.NewLevel(int(litem["level"].(float64))),
			level.NewTick(litem["tick"].(string)))

		if err != nil {
			return nil, err
		}
		if b, ok := litem["check"]; ok {
			cfg.Check = b.(bool)
		}
	}

	result.Redis.SendStatsEnabled = false
	if r, ok := res["redis"]; ok {
		if m, ok := r.(map[string]interface{}); ok {
			result.Redis.Enabled = nvl(m["enabled"], false).(bool)
			result.Redis.Addr = nvl(m["address"], "localhost:6379").(string)
			result.Redis.Password = nvl(m["password"], "").(string)
			result.Redis.DB = int(nvl(m["db"], 0).(float64))
			result.Redis.Queue = int(nvl(m["queue"], 100).(float64))
			result.Redis.Timeout = int(nvl(m["timeout"], 1000).(float64))
			if v, ok := m["periodstats"]; ok {
				result.Redis.PeriodStats, err = time.ParseDuration(nvl(v, "24h").(string))
				if err != nil {
					return nil, err
				}
				result.Redis.SendStatsEnabled = true
				//fmt.Println(result.Redis.PeriodStats.String())
			}
		}
	}
	return result, nil
}

//Check config file
func (c *Config) Check() error {
	if !fileutils.Exists(c.PathToNbackup) {
		return ErrNbackupNotExists
	}
	if !fileutils.Exists(c.Pathtogfix) {
		return ErrGfixNotExists
	}
	if !fileutils.Exists(c.Physicalpathdb) {
		return ErrPhysicalNotExists
	}
	if !fileutils.Exists(c.PathToBackupFolder) {
		return ErrFolderBackupNotExists
	}
	if f, e := os.Stat(c.PathToBackupFolder); e != nil && (os.IsNotExist(e) || !f.IsDir()) {
		return ErrFolderBackupNotExists
	}
	if c.LevelsConfig.Count() == 0 {
		return ErrConfigLevel
	}
	if c.AliasDb == "" {
		return ErrAliasDBNotExists
	}
	return nil
}

//String it Stringer
func (c *Config) String() string {
	var buffer bytes.Buffer
	s := fmt.Sprintf("Config: %s\n", c.file) +
		fmt.Sprintf("Database: name: %q, alias %q, path %q\n", c.NameBase, c.AliasDb, c.Physicalpathdb) +
		fmt.Sprintf("Backup Folder: %s\n", c.PathToBackupFolder) +
		fmt.Sprintf("SMTP Server: %s\n", c.SMTPServer) +
		fmt.Sprintf("Schedule backup:%s\n", c.LevelsConfig.Schedule())
	if _, err := buffer.WriteString(s); err != nil {
		panic(err)
	}
	if c.Redis.Enabled {
		buffer.WriteString(fmt.Sprintf("Send statistics to Redis: %s DB:%d\n", c.Redis.Addr, c.Redis.DB))
		buffer.WriteString(fmt.Sprintf("Redis send timeout: %d (msec),  max queue: %d\n", c.Redis.Timeout, c.Redis.Queue))
	}
	return buffer.String()
}
