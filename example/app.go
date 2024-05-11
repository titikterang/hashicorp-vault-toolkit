package main

import (
	"encoding/json"
	"fmt"
	vault "github.com/titikterang/hashicorp-vault-toolkit/pkg/vault"
	"log"
)

type VaultData struct {
	Data struct {
		RawData YourLocalConfig `json:"data"`
	} `json:"data"`
}

type YourLocalConfig struct {
	SomeSecretUser     string `json:"secret_user"`
	SomeSecretPassword string `json:"secret_password"`
}

/*
for example, bellow is your secret json
{
	"secret_user":"super-secret-username",
	"secret_password":"superUnbreakbleP455w012d"
}
*/

func main() {
	vaultConfig := vault.Config{
		VaultHost:  "https://0.0.0.0:8200",
		VaultToken: "dev-secret-token",
	}
	var SecretPath = "your-secret-path"

	client, err := vault.InitClient(&vaultConfig)
	if err != nil {
		log.Fatalf("err InitClient : %#v\n", err)
	}

	secret, err := client.GetVaultRawSecret(SecretPath)
	if err != nil {
		log.Fatalf("err GetVaultSecret : %#v\n", err)
	}

	var result VaultData
	json.Unmarshal(secret, &result)
	fmt.Printf("\nval %s", result.Data.RawData.SomeSecretUser)
}
