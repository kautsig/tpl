package world

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
)

func (w *World) Vault() *Vault {
	if w.vault != nil {
		return w.vault
	}
	var client *vault.Client
	var err error
	var token string
	vaultConfig := &vault.Config{}
	if err := vaultConfig.ReadEnvironment(); err != nil {
		if w.logger != nil {
			w.logger.Warnf("Failed to read Vault configuration: %s", err.Error())
		}
	}
	client, err = vault.NewClient(vaultConfig)
	if err == nil {
		token = os.Getenv("VAULT_TOKEN")
		if w.logger != nil && token == "" {
			w.logger.Warnf("VAULT_TOKEN not set. Vault not available.")
		} else {
			client.SetToken(token)
		}
	} else {
		if w.logger != nil && token == "" {
			w.logger.Warnf("Failed to create Vault client: %s", err.Error())
		}
	}

	w.vault = &Vault{
		client:     client,
		err:        err,
		logger:     w.logger,
		KeyMapping: make(map[string]string),
	}
	return w.vault
}

type Vault struct {
	logger     *logrus.Logger
	client     *vault.Client
	err        error
	Prefix     string
	KeyMapping map[string]string
}

func (v *Vault) Secret(path, field string) string {
	if v.client == nil {
		return ""
	}
	prefixPath := fmt.Sprintf("%s%s", v.Prefix, path)
	mapped, ok := v.KeyMapping[prefixPath]
	if !ok {
		mapped = path
	}
	sec, err := v.client.Logical().Read(mapped)
	if err != nil {
		if v.logger != nil {
			v.logger.WithError(err).Errorf("Failed to access Vault path %s", mapped)
		}
		return ""
	}
	if sec == nil {
		if v.logger != nil {
			v.logger.Errorf("Vault path %s contained no secret", mapped)
		}
		return ""
	}
	raw, ok := sec.Data[field]
	if !ok {
		if v.logger != nil {
			v.logger.Errorf("%s has no field named '%s'", mapped, field)
		}
		return ""
	}
	return fmt.Sprintf("%s", raw)
}
