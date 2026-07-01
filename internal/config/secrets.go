package config

import (
	"gopkg.in/yaml.v3"

	"github.com/getsops/sops/v3/decrypt"
)

type (
	Secrets struct {
		AdminPassword string `yaml:"admin_password"`
		JwtSecret     string `yaml:"jwt_secret"`
	}
)

func LoadSecretsFromSopsFile(sopsSecretFilePath string) (*Secrets, error) {
	var secrets Secrets
	config, err := decrypt.File(sopsSecretFilePath, "yaml")
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(config, &secrets); err != nil {
		return nil, err
	}

	return &secrets, nil
}
