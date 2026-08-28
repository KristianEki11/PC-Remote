package tlsutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// EnsureCertificates loads existing TLS certificates or generates new ones.
// Returns paths to cert/key files and the SHA-256 fingerprint of the certificate.
// The fingerprint is embedded in the QR code so the Flutter app can pin the TLS cert.
func EnsureCertificates(certDir string) (certFile, keyFile, fingerprint string, err error) {
	certFile = filepath.Join(certDir, "cert.pem")
	keyFile = filepath.Join(certDir, "key.pem")

	// Check if certs already exist and are still valid
	if fileExists(certFile) && fileExists(keyFile) {
		fp, validErr := loadAndValidateFingerprint(certFile)
		if validErr == nil {
			slog.Info("TLS certificates loaded", "cert", certFile, "fingerprint", fp[:16]+"...")
			return certFile, keyFile, fp, nil
		}
		slog.Warn("Existing TLS certificate invalid, regenerating", "reason", validErr)
	}

	// Generate new certificates
	slog.Info("Generating new self-signed TLS certificate", "dir", certDir)
	return generateCertificates(certDir)
}

func generateCertificates(certDir string) (certFile, keyFile, fingerprint string, err error) {
	if err := os.MkdirAll(certDir, 0700); err != nil {
		return "", "", "", fmt.Errorf("failed to create cert directory: %w", err)
	}

	// Generate ECDSA P-256 private key (fast, small, secure)
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate key: %w", err)
	}

	// Random serial number
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate serial: %w", err)
	}

	now := time.Now()
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "PCRemote Server",
			Organization: []string{"PCRemote"},
		},
		NotBefore: now.Add(-1 * time.Hour), // slight backdate for clock skew
		NotAfter:  now.Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		// SANs — allow connections via IP addresses and localhost
		IPAddresses: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("::1"),
		},
		DNSNames: []string{"localhost", "pcremote.local"},
	}

	// Discover and add all local network IPs as Subject Alternative Names
	// so the cert is valid when accessed via any of the host's IPs.
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				template.IPAddresses = append(template.IPAddresses, ipNet.IP)
			}
		}
	}

	// Self-sign
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create certificate: %w", err)
	}

	// Write certificate PEM
	certFile = filepath.Join(certDir, "cert.pem")
	certOut, err := os.OpenFile(certFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to write cert: %w", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		certOut.Close()
		return "", "", "", fmt.Errorf("failed to encode cert PEM: %w", err)
	}
	certOut.Close()

	// Write private key PEM
	keyFile = filepath.Join(certDir, "key.pem")
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to write key: %w", err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		keyOut.Close()
		return "", "", "", fmt.Errorf("failed to marshal key: %w", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}); err != nil {
		keyOut.Close()
		return "", "", "", fmt.Errorf("failed to encode key PEM: %w", err)
	}
	keyOut.Close()

	fingerprint = calculateFingerprint(certDER)
	slog.Info("TLS certificate generated",
		"cert", certFile,
		"valid_until", template.NotAfter.Format("2006-01-02"),
		"fingerprint", fingerprint[:16]+"...",
		"san_ips", len(template.IPAddresses),
	)

	return certFile, keyFile, fingerprint, nil
}

// loadAndValidateFingerprint reads an existing certificate and returns its fingerprint.
// Returns an error if the certificate is expired or malformed.
func loadAndValidateFingerprint(certFile string) (string, error) {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Check expiry with 30-day grace period for regeneration
	if time.Now().After(cert.NotAfter.Add(-30 * 24 * time.Hour)) {
		return "", fmt.Errorf("certificate expires within 30 days (expires: %s)", cert.NotAfter.Format("2006-01-02"))
	}

	return calculateFingerprint(block.Bytes), nil
}

// calculateFingerprint returns the SHA-256 fingerprint of a DER-encoded certificate.
func calculateFingerprint(certDER []byte) string {
	hash := sha256.Sum256(certDER)
	return hex.EncodeToString(hash[:])
}

// InterfaceInfo describes a network interface on the host machine.
type InterfaceInfo struct {
	Name      string `json:"name"`
	IP        string `json:"ip"`
	IsPrimary bool   `json:"is_primary"`
}

// GetLocalIP returns the primary active LAN IPv4 address.
// It uses an outbound UDP probe to query the OS routing table for the real active network adapter,
// filtering out non-routable link-local APIPA addresses (169.254.x.x) and loopbacks.
func GetLocalIP() string {
	// 1. Try OS routing table via UDP probe (does not send packets, returns the default route interface)
	conn, err := net.DialTimeout("udp", "8.8.8.8:80", 500*time.Millisecond)
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		if localAddr.IP != nil {
			ip4 := localAddr.IP.To4()
			if ip4 != nil && !ip4.IsLoopback() && !ip4.IsLinkLocalUnicast() {
				return ip4.String()
			}
		}
	}

	// 2. Fallback: inspect all network interfaces with heuristic ranking
	interfaces := GetAllLANInterfaces()
	if len(interfaces) > 0 {
		return interfaces[0].IP
	}

	return "127.0.0.1"
}

// GetAllLANInterfaces returns all usable, non-loopback, non-APIPA IPv4 network interfaces.
// The primary/most likely Wi-Fi/Ethernet interface is placed first.
func GetAllLANInterfaces() []InterfaceInfo {
	primaryIP := ""
	conn, err := net.DialTimeout("udp", "8.8.8.8:80", 300*time.Millisecond)
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		if localAddr.IP != nil {
			if ip4 := localAddr.IP.To4(); ip4 != nil && !ip4.IsLoopback() && !ip4.IsLinkLocalUnicast() {
				primaryIP = ip4.String()
			}
		}
	}

	var results []InterfaceInfo
	ifaces, err := net.Interfaces()
	if err != nil {
		return results
	}

	for _, iface := range ifaces {
		// Skip down interfaces and loopbacks
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP.To4()
			case *net.IPAddr:
				ip = v.IP.To4()
			}

			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsMulticast() {
				continue
			}

			ipStr := ip.String()
			isPrimary := (ipStr == primaryIP)

			// Clean interface name for display
			name := iface.Name
			if name == "" {
				name = "Network Adapter"
			}

			info := InterfaceInfo{
				Name:      name,
				IP:        ipStr,
				IsPrimary: isPrimary,
			}

			// Prepend primary interface to front of list
			if isPrimary {
				results = append([]InterfaceInfo{info}, results...)
			} else {
				results = append(results, info)
			}
		}
	}

	return results
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
