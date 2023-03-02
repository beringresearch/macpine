package utils

import (
	"fmt"
	"os"
	"strings"
)

type CredentialType int

const (
	PwdCred CredentialType = iota
	HostCred
)

type Credential struct {
	CRType CredentialType
	CR     string
}

/*
credential backends: raw, env, ssh-agent
* if you want to store credentials in the macOS keychain, configure your SSH agent to use the keychain
* - raw:         "raw::password" (password is a string directly after "raw::" prefix)
* - env:         "env::PASS_VAR" (password is stored in environment variable $PASS_VAR)
* - ssh-agent:   "ssh::HOST"     (credential is stored in ssh-agent and configured for use with host HOST in the ssh config)

* `ssh-agent` is the most secure by far, as it allows certificate-based authentication rather than using passwords.
* If `ssh-agent` is configured and working with certificate-based authentication, `PasswordAuthentication no` can be
* set in `/etc/ssh/sshd_config` to significantly harden the VM.
*
* `env` is more secure than `raw`, and may be useful for automation using macpine on systems where configuring `ssh-agent`
* is inconvenient.
*/
func GetCredential(config string) (Credential, error) {
	var cred Credential
	var err error = nil
	if strings.HasPrefix(config, "raw::") {
		cred.CR = strings.TrimPrefix(config, "raw::")
		cred.CRType = PwdCred
	} else if strings.HasPrefix(config, "env::") {
		envvar := strings.TrimPrefix(config, "env::")
		val, ok := os.LookupEnv(envvar)
		if ok {
			cred.CR = val
			cred.CRType = PwdCred
		} else {
			err = fmt.Errorf("config.yaml specifies environment variable credential but variable is not set.")
		}
	} else if strings.HasPrefix(config, "ssh::") {
		cred.CR = strings.TrimPrefix(config, "ssh::")
		cred.CRType = HostCred
	} else {
		// likely a legacy config file with a raw password and no prefix
		cred.CR = config
		cred.CRType = PwdCred
	}
	return cred, err
}
