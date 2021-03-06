package config

//Config represents the global configuration of the system, and the JSON structure on disk.
type Config struct {
	Name                string //canonical name to help identify this server
	DataPath            string
	DeploymentPath      string
	OverrideStatusColor string
	Path                string   `json:"-"` //path used to represent where the file is currently stored.
	RunSentryInterval   int      //Seconds between executions of the runsentry.
	TLS                 struct { //Relative file addresses of the .pem files needed for TLS.
		Enabled       bool
		ForceRedirect bool //If set, all HTTPPROXY requests for apps must go over HTTPS. HTTP traffic is redirected.
		Listener      string
		Certs         []struct {
			PrivateKey string //PEM format recommended.
			Cert       string //PEM format required. You may need to contain the whole chain in the cert as well - to do this: cat mycert.pem CA.pem > cert_for_pushtart.pem
		}
	}

	Web struct { //Details needed to get the website part working.
		Enabled       bool
		DefaultDomain string //Domain should be in the form example.com
		Listener      string //Address:port (address can be omitted) where the HTTPS listener will bind.
		DomainProxies map[string]DomainProxy
		LogAllProxies bool
	}

	APIKeys []struct {
		Service string
		Key     string
	}

	SSH struct {
		PubPEM   string
		PrivPEM  string
		Listener string
	}

	DNS struct {
		Enabled         bool
		Listener        string
		AllowForwarding bool
		LookupCacheSize int
		ARecord         map[string]ARecord
		AAAARecord      map[string]ARecord
	}

	Users map[string]User
	Tarts map[string]Tart
}

//DomainProxy represents a reverse proxy for requests received on a specific domain, to a specific host/port.
type DomainProxy struct {
	TargetHost   string
	TargetPort   int
	TargetScheme string
	AuthRules    []AuthorizationRule
}

//AuthorizationRule is a ALLOW/DENY rule for a specific domain
type AuthorizationRule struct {
	RuleType string
	Username string
}

//ARecord represents a response that could be served to a DNS query of type A.
type ARecord struct {
	Address string
	TTL     uint32
}

//User represents an account which has access to the system.
type User struct {
	Name             string
	Password         string
	AllowSSHPassword bool
	SSHPubKey        string
}

//Tart stores information for tarts which are stored in the system.
type Tart struct {
	PushURL          string
	Name             string
	Owners           []string
	IsRunning        bool
	LogStdout        bool
	PID              int
	Env              []string
	RestartOnStop    bool
	RestartDelaySecs int
	LastHash         string
	LastGitMessage   string
}
