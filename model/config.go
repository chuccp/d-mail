package model

import "github.com/chuccp/go-web-frame/config"

type Config struct {
	Core   *CoreConfig   `json:"core"`
	Manage *ManageConfig `json:"manage"`
	Api    *ApiConfig    `json:"api"`
	Sqlite *SqliteConfig `json:"sqlite"`
	Mysql  *MysqlConfig  `json:"mysql"`
}

func DefaultConfig() *Config {
	return &Config{
		Core:   DefaultCoreConfig(),
		Manage: DefaultManageConfig(),
		Api:    DefaultApiConfig(),
		Sqlite: DefaultSqliteConfig(),
		Mysql:  DefaultMysqlConfig(),
	}
}

func GetConfig(config config.IConfig) (*Config, error) {
	cfg := DefaultConfig()
	// Pass the *Config itself, not &cfg: the decoder only applies its per-field
	// weakly-typed conversions when the target is a struct. A **Config falls back to
	// encoding/json, which cannot coerce the strings an INI file yields into
	// bool/int fields ("true" -> bool, "12566" -> int).
	err := config.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// CoreState reads the core section through the typed config.
//
// Do not reach for IConfig.GetBoolOrDefault("core.init") instead: values loaded from an
// INI file are strings, and that getter only accepts a real bool, so it silently reports
// the default. Every setup guard and the startup database attach depend on this.
func CoreState(cfg config.IConfig) *CoreConfig {
	c, err := GetConfig(cfg)
	if err != nil || c.Core == nil {
		return DefaultCoreConfig()
	}
	return c.Core
}

// MaskSecrets blanks credential fields so the config can be returned to clients.
// Callers that persist config treat an empty password as "keep the existing one".
func (cfg *Config) MaskSecrets() *Config {
	if cfg != nil && cfg.Mysql != nil {
		cfg.Mysql.Password = ""
	}
	return cfg
}

type CoreConfig struct {
	Init      bool   `json:"init"`
	DbInit    bool   `json:"dbInit"`
	CachePath string `json:"cachePath"`
	DbType    string `json:"dbType"`
	LogLevel  string `json:"logLevel"`
	IsDocker  bool   `json:"isDocker"`
	Debug     bool   `json:"debug"`
}

func DefaultCoreConfig() *CoreConfig {
	return &CoreConfig{
		Init:      false,
		DbInit:    false,
		CachePath: ".cache",
		DbType:    "sqlite",
		LogLevel:  "info",
		IsDocker:  false,
		Debug:     false,
	}
}

type ManageConfig struct {
	Port    int    `json:"port"`
	WebPath string `json:"webPath"`
}

var ManagePort = 12566
var ApiPort = 12567

func DefaultManageConfig() *ManageConfig {
	return &ManageConfig{
		Port:    12566,
		WebPath: "web",
	}
}

type ApiConfig struct {
	Port int `json:"port"`
}

func DefaultApiConfig() *ApiConfig {
	return &ApiConfig{
		Port: 12567,
	}
}

type SqliteConfig struct {
	Filename string `json:"filename"`
}

func DefaultSqliteConfig() *SqliteConfig {
	return &SqliteConfig{
		Filename: "d-mail.db",
	}
}

type MysqlConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Dbname   string `json:"dbname"`
	Username string `json:"username"`
	Password string `json:"password"`
	Charset  string `json:"charset"`
}

func DefaultMysqlConfig() *MysqlConfig {
	return &MysqlConfig{
		Host:     "",
		Port:     3306,
		Dbname:   "d-main",
		Username: "",
		Password: "",
		Charset:  "utf8",
	}
}

type System struct {
	HasInit   bool `json:"hasInit"`
	HasDbInit bool `json:"hasDbInit"`
	HasAdmin  bool `json:"hasAdmin"`
	HasLogin  bool `json:"hasLogin"`
	IsDocker  bool `json:"isDocker"`
	// Populated only when the request carries a valid session. The session itself is an
	// HttpOnly cookie the client cannot read, so the UI needs these to render the
	// username and decide whether to show admin-only navigation.
	Username string `json:"username"`
	IsAdmin  bool   `json:"isAdmin"`
}
