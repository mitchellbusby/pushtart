package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"pushtart/constants"
	"pushtart/logging"
	"strconv"

	"golang.org/x/crypto/ssh"
)

func Generate(path string) (err error) {
	logging.Info("config", "Now generating default config to: "+path)

	if gConfig == nil {
		gConfig = &Config{
			Name: "pushtart",
			Path: path,
		}
	}

	if gConfig.Ssh.PrivPEM == "" {
		gConfig.Ssh.PubPEM, gConfig.Ssh.PrivPEM, err = MakeSSHKeyPair()
	}

	if gConfig.Ssh.Listener == "" {
		gConfig.Ssh.Listener = "0.0.0.0:2022"
	}

	return writeConfig()
}

// MakeSSHKeyPair make a pair of public and private keys for SSH access.
// Public key is encoded in the format for inclusion in an OpenSSH authorized_keys file.
// Private Key generated is PEM encoded
// Source: http://stackoverflow.com/questions/21151714/go-generate-an-ssh-public-key
func MakeSSHKeyPair() (pubKey, privKey string, err error) {
	logging.Info("config-generate", "Now generating SSH private key.")
	logging.Info("config-generate", "Key scheme: RSA. Key size: "+strconv.Itoa(constants.RsaKeySize))
	privateKey, err := rsa.GenerateKey(rand.Reader, constants.RsaKeySize)
	if err != nil {
		return "", "", err
	}

	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	privKey = string(pem.EncodeToMemory(privateKeyPEM))

	// generate and write public key
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	pubKey = string(ssh.MarshalAuthorizedKey(pub))
	return
}
