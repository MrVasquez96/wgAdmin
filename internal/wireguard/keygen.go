package wireguard

import (
	"os/exec"
	"strings"
)

// GeneratePrivateKey generates a new WireGuard private key
func GeneratePrivateKey() (string, error) {
	cmd := exec.Command("wg", "genkey")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// DerivePublicKey derives public key from private key
func DerivePublicKey(privateKey string) (string, error) {
	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(privateKey)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GenerateKeyPair generates a new private/public key pair
func GenerateKeyPair() (privateKey, publicKey string, err error) {
	privateKey, err = GeneratePrivateKey()
	if err != nil {
		return "", "", err
	}

	publicKey, err = DerivePublicKey(privateKey)
	if err != nil {
		return "", "", err
	}

	return privateKey, publicKey, nil
}

// GeneratePresharedKey generates a preshared key for additional security
func GeneratePresharedKey() (string, error) {
	cmd := exec.Command("wg", "genpsk")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
