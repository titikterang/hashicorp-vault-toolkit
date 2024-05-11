package vault

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	circuit "github.com/eapache/go-resiliency/breaker"
	vault "github.com/hashicorp/vault/api"
	"io/ioutil"
	"net/http"
	"time"
)

func InitClient(cfg *Config) (*VaultAPI, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	if cfg.APIConfig == nil {
		cfg.APIConfig = &APIConfig{
			LimitBreakerErrorThreshold:   BreakerErrorThreshold,
			LimitBreakerSuccessThreshold: BreakerSuccessThreshold,
			LimitBreakerTimeout:          BreakerTimeout,
		}
	}

	if cfg.HttpClientConfig == nil {
		cfg.HttpClientConfig = &HttpClientConfig{
			LimitPoolClientTimeoutSeconds:            PoolClientTimeoutSeconds,
			LimitPoolTransportMaxIdleConns:           PoolTransportMaxIdleConns,
			LimitPoolTransportMaxIdleConnsPerHost:    PoolTransportMaxIdleConnsPerHost,
			LimitPoolTransportIdleConnTimeoutSeconds: PoolTransportIdleConnTimeoutSeconds,
		}
	}

	vaultPool := &http.Client{
		Timeout: cfg.HttpClientConfig.LimitPoolClientTimeoutSeconds * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.HttpClientConfig.LimitPoolTransportMaxIdleConns,
			MaxIdleConnsPerHost: cfg.HttpClientConfig.LimitPoolTransportMaxIdleConnsPerHost,
			IdleConnTimeout:     cfg.HttpClientConfig.LimitPoolTransportIdleConnTimeoutSeconds * time.Second,
		},
	}

	// setup vault api http client
	return &VaultAPI{
		CircuitBreaker: circuit.New(cfg.APIConfig.LimitBreakerErrorThreshold, cfg.APIConfig.LimitBreakerSuccessThreshold, cfg.APIConfig.HttpClientPoolTimeoutSec*time.Second),
		Config:         cfg,
		Client:         vaultPool,
	}, nil
}

func (v *VaultAPI) GetVaultSecret(path string, useV1 bool) (respData VaultResponseData, err error) {
	var V2Resp VaultKV2RespData
	err = v.CircuitBreaker.Run(func() error {
		req, errCB := http.NewRequest("GET", fmt.Sprintf("%s/v1/%s", v.Config.VaultHost, path), nil)
		if errCB != nil {
			return errCB
		}

		req.Header.Set("X-Vault-Token", v.Config.VaultToken)
		resp, errCB := v.Client.Do(req)
		if errCB != nil {
			return errCB
		}

		if resp.StatusCode != http.StatusOK {
			return errors.New(http.StatusText(resp.StatusCode))
		}

		defer resp.Body.Close()
		body, errCB := ioutil.ReadAll(resp.Body)
		if errCB != nil {
			return errCB
		}

		if useV1 {
			errCB = json.Unmarshal(body, &respData)
			if errCB != nil {
				return errCB
			}
		} else {
			errCB = json.Unmarshal(body, &V2Resp)
			if errCB != nil {
				return errCB
			}
			respData = V2Resp.Data
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("code %d, resp data : %+v", resp.StatusCode, string(body))
		}

		return nil
	})

	return
}

func (v *VaultAPI) GetVaultRawSecret(path string) (respData RawFileVaultResponseData, err error) {
	err = v.CircuitBreaker.Run(func() error {
		req, errCB := http.NewRequest("GET", fmt.Sprintf("%s/v1/%s", v.Config.VaultHost, path), nil)
		if errCB != nil {
			return errCB
		}

		req.Header.Set("X-Vault-Token", v.Config.VaultToken)
		resp, errCB := v.Client.Do(req)
		if errCB != nil {
			return errCB
		}

		if resp.StatusCode != http.StatusOK {
			return errors.New(http.StatusText(resp.StatusCode))
		}

		defer resp.Body.Close()
		body, errCB := ioutil.ReadAll(resp.Body)
		if errCB != nil {
			return errCB
		}

		errCB = json.Unmarshal(body, &respData)
		if errCB != nil {
			return errCB
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("code %d, GetVaultRawSecret resp data : %+v", resp.StatusCode, string(body))
		}

		return nil
	})

	return
}

func (v *VaultAPI) GetKVMapString(path string) (map[string]interface{}, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), v.Config.HttpClientConfig.LimitPoolClientTimeoutSeconds*time.Second)
	defer cancelFunc()

	config := vault.DefaultConfig()
	config.Address = v.Config.VaultHost
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}

	// Authenticate
	client.SetToken(v.Config.VaultToken)
	// Read a secret from the default mount path for KV v2 in dev mode, "secret"
	secret, err := client.KVv2("kv").Get(ctx, v.Config.SecretPath)
	if err != nil {
		return nil, err
	}

	return secret.Data, nil
}
