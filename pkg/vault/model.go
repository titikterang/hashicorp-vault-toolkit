package vault

import (
	circuit "github.com/eapache/go-resiliency/breaker"
	"net/http"
	"time"
)

const (
	PoolClientTimeoutSeconds            = 5
	PoolTransportMaxIdleConns           = 100
	PoolTransportMaxIdleConnsPerHost    = 2
	PoolTransportIdleConnTimeoutSeconds = 90

	BreakerTimeout          = 5
	BreakerErrorThreshold   = 10
	BreakerSuccessThreshold = 1
)

type Config struct {
	VaultHost        string
	VaultToken       string
	APIConfig        *APIConfig
	HttpClientConfig *HttpClientConfig
}

type HttpClientConfig struct {
	LimitPoolClientTimeoutSeconds            time.Duration
	LimitPoolTransportIdleConnTimeoutSeconds time.Duration
	LimitPoolTransportMaxIdleConns           int
	LimitPoolTransportMaxIdleConnsPerHost    int
}

type APIConfig struct {
	LimitBreakerErrorThreshold   int
	LimitBreakerSuccessThreshold int
	LimitBreakerTimeout          int
	HttpClientPoolTimeoutSec     time.Duration
}

type VaultAPI struct {
	Config         *Config
	CircuitBreaker *circuit.Breaker
	Client         *http.Client
}
