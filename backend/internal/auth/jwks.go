package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"sync"
	"time"
)

type JWKSCache struct {
	url        string
	ttl        time.Duration
	mu         sync.Mutex
	lastFetch  time.Time
	publicKeys map[string]*rsa.PublicKey
}

type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
	Use string `json:"use"`
}

func NewJWKSCache(url string, ttl time.Duration) *JWKSCache {
	return &JWKSCache{
		url:        url,
		ttl:        ttl,
		publicKeys: map[string]*rsa.PublicKey{},
	}
}

func (c *JWKSCache) GetKey(kid string) (*rsa.PublicKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Since(c.lastFetch) > c.ttl || len(c.publicKeys) == 0 {
		if err := c.refresh(); err != nil {
			return nil, err
		}
	}

	key, ok := c.publicKeys[kid]
	if !ok {
		if err := c.refresh(); err != nil {
			return nil, err
		}
		key, ok = c.publicKeys[kid]
		if !ok {
			return nil, errors.New("jwks: key not found")
		}
	}
	return key, nil
}

func (c *JWKSCache) refresh() error {
	req, err := http.NewRequest(http.MethodGet, c.url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var payload jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}

	keys := map[string]*rsa.PublicKey{}
	for _, key := range payload.Keys {
		if key.Kty != "RSA" || key.Kid == "" {
			continue
		}
		pub, err := jwkToPublicKey(key.N, key.E)
		if err != nil {
			continue
		}
		keys[key.Kid] = pub
	}

	if len(keys) == 0 {
		return errors.New("jwks: no keys parsed")
	}

	c.publicKeys = keys
	c.lastFetch = time.Now()
	return nil
}

func jwkToPublicKey(n string, e string) (*rsa.PublicKey, error) {
	nb, err := base64.RawURLEncoding.DecodeString(n)
	if err != nil {
		return nil, err
	}
	eb, err := base64.RawURLEncoding.DecodeString(e)
	if err != nil {
		return nil, err
	}

	modulus := new(big.Int).SetBytes(nb)
	exponent := new(big.Int).SetBytes(eb)
	if exponent.BitLen() > 31 {
		return nil, errors.New("jwks: exponent too large")
	}

	return &rsa.PublicKey{
		N: modulus,
		E: int(exponent.Int64()),
	}, nil
}
