package loadcreds

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/kubernetes"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/credentialhelper"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/giturl"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/jenkins-x/jx-helpers/v3/pkg/homedir"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
)

const (
	BootSecretName    = "jx-boot"
	OperatorNamespace = "jx-git-operator"
)

// Credentials the loaded credentials
type Credentials struct {
	ServerURL string
	Username  string
	Password  string
	Token     string
}

// GetServerCredentials returns the credentials for the given server URL or an empty struct
func GetServerCredentials(credentials []Credentials, serverURL string) Credentials {
	for i := range credentials {
		c := credentials[i]
		if c.ServerURL == serverURL {
			return c
		}
	}
	return Credentials{}
}

// GitCredentialsFile returns the location of the git credentials file
func GitCredentialsFile() (string, error) {
	homeDir := homedir.HomeDir()
	cfgHome := os.Getenv("XDG_CONFIG_HOME")
	if cfgHome == "" {
		cfgHome = homeDir
	}
	if cfgHome == "" {
		cfgHome = "."
	}
	paths := []string{
		filepath.Join(cfgHome, "git", "credentials"),
		filepath.Join(cfgHome, ".git-credentials"),
	}
	if homeDir != "" && homeDir != cfgHome {
		paths = append(paths,
			filepath.Join(homeDir, "git", "credentials"),
			filepath.Join(homeDir, ".git-credentials"),
		)
	}
	paths = append(paths,
		filepath.Join("git", "credentials"),
		filepath.Join(".git-credentials"),
	)
	for _, path := range paths {
		exists, err := files.FileExists(path)
		if err != nil {
			return path, errors.Wrapf(err, "failed to check if git credentials file exists %s", path)
		}
		if exists {
			return path, nil
		}
	}

	// lets return the default name we think should be used....
	return filepath.Join(cfgHome, "git", "credentials"), nil
}

// LoadGitCredentialsFile loads the git credentials from the `git/credentials` file
// in `$XDG_CONFIG_HOME/git/credentials` or in the `~/git/credentials` directory
func LoadGitCredential() ([]Credentials, error) {
	fileName, err := GitCredentialsFile()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find git credentials file")
	}
	if fileName == "" {
		return nil, nil
	}
	data, _, err := LoadGitCredentialsFile(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load credential file")
	}
	return data, nil
}

// loadGitCredentialsAuthFile loads the git credentials file
func LoadGitCredentialsFile(fileName string) ([]Credentials, bool, error) {
	pMask := "****"
	log.Logger().Debugf("loading git credentials file %s", termcolor.ColorInfo(fileName))

	exists, err := files.FileExists(fileName)
	if err != nil {
		return nil, false, errors.Wrapf(err, "failed to check for file %s", fileName)
	}
	if !exists {
		return nil, false, nil
	}

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, true, errors.Wrapf(err, "failed to load git credentials file %s", fileName)
	}

	var answer []Credentials
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		u, err := url.Parse(line)
		if err != nil {
			log.Logger().Warnf("ignoring invalid line in git credentials file: %s error: %s", fileName, err.Error())
			continue
		}

		user := u.User
		username := user.Username()
		password, _ := user.Password()
		if username == "" {
			log.Logger().Warnf("ignoring missing user name in git credentials file: %s URL: %s", fileName, strings.ReplaceAll(line, password, pMask))
			continue
		}
		if password == "" {
			log.Logger().Warnf("ignoring missing password in git credentials file: %s URL: %s", fileName, line)
			continue
		}
		u.User = nil
		config := Credentials{}
		config.ServerURL = u.String()
		config.Username = username
		config.Password = password
		answer = append(answer, config)
	}
	return answer, true, nil
}

// FindOperatorCredentials detects the git operator secret so we have default credentials
func FindOperatorCredentials() (credentialhelper.GitCredential, error) {
	var client kubernetes.Interface
	var credential credentialhelper.GitCredential
	var err error
	client, ns, err := kube.LazyCreateKubeClientAndNamespace(client, "")
	if err != nil {
		return credential, errors.Wrapf(err, "failed to create kube client")
	}
	secret, err := client.CoreV1().Secrets(ns).Get(context.TODO(), BootSecretName, metav1.GetOptions{})
	if err != nil && ns != OperatorNamespace {
		var err2 error
		secret, err2 = client.CoreV1().Secrets(OperatorNamespace).Get(context.TODO(), BootSecretName, metav1.GetOptions{})
		if err2 == nil {
			err = nil
		}
	}
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Logger().Warnf("could not find secret %s in namespace %s", BootSecretName, ns)
			return credential, nil
		}
		return credential, errors.Wrapf(err, "failed to find Secret %s in namespace %s", BootSecretName, ns)
	}
	data := secret.Data
	if data == nil {
		return credential, errors.Wrapf(err, "failed to find data in secret %s", BootSecretName)
	}

	gitURL := string(data["url"])
	if gitURL == "" {
		log.Logger().Warnf("secret %s in namespace %s does not have a url entry", BootSecretName, ns)
		return credential, nil
	}
	// lets convert the git URL into a provider URL
	gitInfo, err := giturl.ParseGitURL(gitURL)
	if err != nil {
		return credential, errors.Wrapf(err, "failed to parse git URL %s", gitURL)
	}
	gitProviderURL := gitInfo.HostURL()
	username := string(data["username"])
	password := string(data["password"])
	credential, err = credentialhelper.CreateGitCredentialFromURL(gitProviderURL, username, password)
	if err != nil {
		return credential, errors.Wrapf(err, "invalid git auth information")
	}
	return credential, nil
}

// FindGitCredentialsFromSecret detects the git secrets using a secret name
func FindGitCredentialsFromSecret(secretName string) (credentialhelper.GitCredential, error) {
	var client kubernetes.Interface
	var credential credentialhelper.GitCredential
	var err error
	client, ns, err := kube.LazyCreateKubeClientAndNamespace(client, "")
	if err != nil {
		return credential, errors.Wrapf(err, "failed to create kube client")
	}
	secret, err := client.CoreV1().Secrets(ns).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return credential, errors.Wrapf(err, "failed to find Secret %s in namespace %s", BootSecretName, ns)
	}

	data := secret.Data
	if data == nil {
		return credential, errors.Wrapf(err, "failed to find data in secret %s", secretName)
	}

	gitURL := string(data["url"])
	if gitURL == "" {
		log.Logger().Warnf("secret %s in namespace %s does not have a url entry", secretName, ns)
		return credential, nil
	}
	// lets convert the git URL into a provider URL
	gitInfo, err := giturl.ParseGitURL(gitURL)
	if err != nil {
		return credential, errors.Wrapf(err, "failed to parse git URL %s", gitURL)
	}
	gitProviderURL := gitInfo.HostURL()
	username := string(data["username"])
	password := string(data["password"])
	credential, err = credentialhelper.CreateGitCredentialFromURL(gitProviderURL, username, password)
	if err != nil {
		return credential, errors.Wrapf(err, "invalid git auth information")
	}
	return credential, nil
}
