package config

import "time"

type RedisConfig struct {
	ConnAddr string        `yaml:"conn_addr" env:"CONN_ADDR" required:"true"`
	Password string        `yaml:"conn_pass" env:"PASSWORD"`
	DB       int           `yaml:"conn_db" env:"DB"`
	Duration time.Duration `yaml:"duration" env:"DURATION" default:"1440m"`
}
