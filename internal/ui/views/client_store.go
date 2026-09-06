package ui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/MrVasquez96/go-wg/wg/config"
)

// A peer's private key never appears in the server's config — WireGuard only
// ever sees its public half — so wgAdmin used to be able to build a client
// config or QR code exactly once: in the session that generated the key pair.
// Restart the app and the peer was unprovisionable forever.
//
// The keys are not actually lost, though: every generated client config is
// written to the client config directory and contains the peer's private key in
// its own [Interface] section. Reading them back is what lets a QR code be
// shown at any point rather than only at creation time.

// clientDirFor is the directory a tunnel's generated client configs live in.
func clientDirFor(root, tunnel string) string {
	return filepath.Join(root, tunnel)
}

// storedPeerKeys maps peer public keys to the private keys held by a tunnel's
// stored client configs. It is empty when the directory does not exist, which
// is the normal case for a tunnel whose peers were all added by public key.
//
// Keys are matched by deriving the public half from each stored private key
// rather than by filename, so renaming a peer does not orphan its key.
func storedPeerKeys(root, tunnel string) map[string]string {
	keys := map[string]string{}
	if root == "" || tunnel == "" {
		return keys
	}

	dir := clientDirFor(root, tunnel)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return keys
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".conf") {
			continue
		}

		cfg, err := config.ParseConfig(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		keys[cfg.Interface.PrivateKey.PublicKey().String()] = cfg.Interface.PrivateKey.String()
	}

	return keys
}

// storedPeerPrivateKey is storedPeerKeys for a single peer. Callers that ask
// about several peers should read the map once instead: this re-reads the
// directory on every call.
func storedPeerPrivateKey(root, tunnel, publicKey string) string {
	if publicKey == "" {
		return ""
	}
	return storedPeerKeys(root, tunnel)[publicKey]
}
