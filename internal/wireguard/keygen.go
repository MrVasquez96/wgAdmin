package wireguard

import (
	"fmt"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// GeneratePrivateKey generates a new WireGuard private key
func GeneratePrivateKey() (string, error) {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate private key: %w", err)
	}
	return key.String(), nil
}

// DerivePublicKey derives public key from private key
func DerivePublicKey(privateKey string) (string, error) {
	key, err := wgtypes.ParseKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}
	return key.PublicKey().String(), nil
}

// GenerateKeyPair generates a new private/public key pair
func GenerateKeyPair() (privateKey, publicKey string, err error) {
	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}
	return privKey.String(), privKey.PublicKey().String(), nil
}

// GeneratePresharedKey generates a preshared key for additional security
func GeneratePresharedKey() (string, error) {
	key, err := wgtypes.GenerateKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate preshared key: %w", err)
	}
	return key.String(), nil
}
