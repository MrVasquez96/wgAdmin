package wireguard

import (
	"encoding/base64"
	"fmt"
	"net"
	"regexp"
	"strings"

	"wgAdmin/internal/models"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateConfig checks a config for common errors
func ValidateConfig(config *models.WireGuardConfig) []error {
	var errs []error

	// Validate Interface section
	if config.PrivateKey == "" {
		errs = append(errs, ValidationError{Field: "PrivateKey", Message: "required"})
	} else if !ValidateKey(config.PrivateKey) {
		errs = append(errs, ValidationError{Field: "PrivateKey", Message: "invalid format"})
	}

	if config.Address == "" {
		errs = append(errs, ValidationError{Field: "Address", Message: "required"})
	} else if !ValidateAddress(config.Address) {
		errs = append(errs, ValidationError{Field: "Address", Message: "invalid CIDR format"})
	}

	if config.ListenPort < 0 || config.ListenPort > 65535 {
		errs = append(errs, ValidationError{Field: "ListenPort", Message: "must be 0-65535"})
	}

	if config.MTU < 0 || config.MTU > 65535 {
		errs = append(errs, ValidationError{Field: "MTU", Message: "must be 0-65535"})
	}

	// Validate Peers
	for i, peer := range config.Peers {
		prefix := fmt.Sprintf("Peer[%d]", i)

		if peer.PublicKey == "" {
			errs = append(errs, ValidationError{Field: prefix + ".PublicKey", Message: "required"})
		} else if !ValidateKey(peer.PublicKey) {
			errs = append(errs, ValidationError{Field: prefix + ".PublicKey", Message: "invalid format"})
		}

		if peer.AllowedIPs == "" {
			errs = append(errs, ValidationError{Field: prefix + ".AllowedIPs", Message: "required"})
		} else if !ValidateAllowedIPs(peer.AllowedIPs) {
			errs = append(errs, ValidationError{Field: prefix + ".AllowedIPs", Message: "invalid CIDR format"})
		}

		if peer.Endpoint != "" && !ValidateEndpoint(peer.Endpoint) {
			errs = append(errs, ValidationError{Field: prefix + ".Endpoint", Message: "invalid format (host:port)"})
		}

		if peer.PersistentKeepalive < 0 || peer.PersistentKeepalive > 65535 {
			errs = append(errs, ValidationError{Field: prefix + ".PersistentKeepalive", Message: "must be 0-65535"})
		}

		if peer.PresharedKey != "" && !ValidateKey(peer.PresharedKey) {
			errs = append(errs, ValidationError{Field: prefix + ".PresharedKey", Message: "invalid format"})
		}
	}

	return errs
}

// ValidateKey checks if a key is valid base64 and correct length (32 bytes = 44 chars base64)
func ValidateKey(key string) bool {
	if len(key) != 44 {
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return false
	}
	return len(decoded) == 32
}

// ValidateAddress checks CIDR notation (e.g., 10.0.0.1/24)
func ValidateAddress(addr string) bool {
	_, _, err := net.ParseCIDR(addr)
	return err == nil
}

// ValidateAllowedIPs checks comma-separated CIDRs
func ValidateAllowedIPs(ips string) bool {
	parts := strings.Split(ips, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, _, err := net.ParseCIDR(part); err != nil {
			return false
		}
	}
	return true
}

// ValidateEndpoint checks host:port format
func ValidateEndpoint(endpoint string) bool {
	// Can be hostname:port or ip:port
	pattern := regexp.MustCompile(`^[a-zA-Z0-9.-]+:\d{1,5}$`)
	if !pattern.MatchString(endpoint) {
		return false
	}
	// Check port range
	parts := strings.Split(endpoint, ":")
	if len(parts) != 2 {
		return false
	}
	var port int
	fmt.Sscanf(parts[1], "%d", &port)
	return port > 0 && port <= 65535
}

// ValidateName checks if a tunnel name is valid
func ValidateName(name string) bool {
	if name == "" || len(name) > 15 {
		return false
	}
	// Only alphanumeric and underscore/hyphen
	pattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return pattern.MatchString(name)
}
