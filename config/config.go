package config

import (
	"fmt"

	"gopkg.in/ini.v1"
)

type Config struct {
	Host              string
	Port              int
	User              string
	Password          string
	File              string
	Workers           int
	ChunkSize         int
	Database          string
	ChannelBufferSize int
	IniFile           string
	Directory         string
	Pattern           string
}

func NewConfig() *Config {
	return &Config{
		Host:              "localhost",
		Port:              3306,
		Workers:           4,
		ChunkSize:         50000,
		ChannelBufferSize: 2000,
		Pattern:           "*-thread*.sql",
	}
}

func (c *Config) LoadIniFile() error {
	if c.IniFile == "" {
		return nil
	}

	cfg, err := ini.Load(c.IniFile)
	if err != nil {
		return fmt.Errorf("failed to load ini file: %v", err)
	}

	section := cfg.Section("go-load")
	if section.HasKey("mysql-user") {
		c.User = section.Key("mysql-user").String()
	}
	if section.HasKey("mysql-password") {
		c.Password = section.Key("mysql-password").String()
	}
	if section.HasKey("mysql-host") {
		c.Host = section.Key("mysql-host").String()
	}

	return nil
}
