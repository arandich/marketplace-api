package config

type QueueConfig struct {
	AWSAccessKeyID     string `yaml:"aws_access_key_id" env:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `yaml:"aws_secret_access_key" env:"AWS_SECRET_ACCESS_KEY"`
	URL                string `yaml:"url" env:"URL"`
	Region             string `yaml:"region" env:"REGION"`
	Workers            int    `yaml:"workers" env:"WORKERS"`
}
