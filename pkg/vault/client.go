// Package vault implements a wrapper around a Vault API client that retrieves
// credentials from the operating system environment.
package vault

import (
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

//  Client initializes a Vault client using the environment variables:
// VAULT_ADDR, VAULT_ROLE_ID, VAULT_SECRET_ID, VAULT_TOKEN.
//
// Because individual tokens have usage limits, we re-authenticate for each new
// client.
func Client() *api.Client {
	vaultCFG := api.DefaultConfig()
	vaultCFG.Address = mustGetenv("VAULT_ADDR")
	var clientToken string
	var client *api.Client

	if clientToken == "" {
		var err error
		client, err = api.NewClient(vaultCFG)
		if err != nil {
			log.WithError(err).Fatal("failed to initialize Vault client")
		}

		switch authType := defaultGetenv("VAULT_AUTHTYPE", "approle"); strings.ToLower(authType) {
		case "approle":
			roleID := mustGetenv("VAULT_ROLE_ID")
			secretID := mustGetenv("VAULT_SECRET_ID")

			secret, err := client.Logical().Write("auth/approle/login", map[string]interface{}{
				"role_id":   roleID,
				"secret_id": secretID,
			})
			if err != nil {
				log.WithError(err).WithField("package", "vault").Fatal("failed to login to Vault with AppRole")
			}
			clientToken = secret.Auth.ClientToken
		case "token":
			clientToken = mustGetenv("VAULT_TOKEN")
		default:
			log.WithField("authType", authType).Fatal("unsuported auth type")
		}
	}

	client.SetToken(clientToken)
	return client
}

func mustGetenv(name string) string {
	env := os.Getenv(name)
	if env == "" {
		log.WithField("env", name).Fatal("required environment variable is unset")
	}
	return env
}

func defaultGetenv(name, defaultName string) string {
	env := os.Getenv(name)
	if env == "" {
		env = defaultName
	}
	return env
}
