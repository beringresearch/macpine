package utils

import (
	"fmt"
	"os"
	"strings"
)

type CredentialType int
const (
   Password CredentialType = iota
   Host
)
type Credential struct {
   CRType CredentialType;
   CR string;
}

/* credential backends: raw, env, ssh-agent
 * if you want to store credentials in the macOS keychain, configure your SSH agent to use the keychain
 * - raw:         "raw::password" (password is a string directly after "pw:" prefix)
 * - env:         "env::PASS_VAR" (password is stored in environment variable $PASS_VAR)
 * - ssh-agent:   "ssh::HOST"     (credential is stored in ssh-agent and configured for use with host HOST in the ssh config)

 * `ssh-agent` is the most secure by far, as it allows certificate-based authentication rather than using passwords.
 * If `ssh-agent` is configured and working, `PasswordAuthentication no` can be set in `/etc/ssh/sshd_config` to significantly
 * harden the VM.
*/
func GetCredential(config string) (Credential, error) {
   var cred Credential
   var err error = nil
   if strings.HasPrefix(config, "raw::") {
      cred.CR = strings.TrimPrefix(config, "raw::")
      cred.CRType = Password
   } else if strings.HasPrefix(config, "env::") {
      envvar := strings.TrimPrefix(config, "env::")
      val, ok := os.LookupEnv(envvar)
      if ok {
         cred.CR = val
         cred.CRType = Password
      } else {
         err = fmt.Errorf("config.yaml specifies environment variable credential but variable is not set.")
      }
   } else if strings.HasPrefix(config, "ssh::") {
      cred.CR = strings.TrimPrefix(config, "ssh::")
      cred.CRType = Host
   } else {
      // likely a legacy config file with a raw password and no prefix
      cred.CR = config
      cred.CRType = Password
   }
   return cred, err
}
