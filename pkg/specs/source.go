package specs

import (
	azure "github.com/masinger/incredible/pkg/azure/specs"
	bitwarden "github.com/masinger/incredible/pkg/bitwarden/specs"
	lastpass "github.com/masinger/incredible/pkg/lastpass/specs"
)

type Source struct {
	Bitwarden           *bitwarden.Source               `yaml:"bitwarden" json:"bitwarden"`
	AzureKeyVaultSecret *azure.KeyVaultSecretSource     `yaml:"azureKeyVaultSecret" json:"azureKeyVaultSecret"`
	LastPass            *lastpass.LastPassAccountSource `yaml:"lastpass" json:"lastpass"`
}
