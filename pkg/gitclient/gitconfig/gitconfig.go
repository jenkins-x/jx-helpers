package gitconfig

import (
	"fmt"
	gitcfg "github.com/go-git/go-git/v5/config"
	"os"
)

// DiscoverUpstreamGitURL discovers the upstream git URL from the given git configuration
func DiscoverUpstreamGitURL(gitConf string, preferUpstream bool) (string, error) {
	cfg, err := parseGitConfig(gitConf)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal %s due to %s", gitConf, err)
	}
	remotes := cfg.Remotes
	if len(remotes) == 0 {
		return "", nil
	}
	names := []string{"origin", "upstream"}
	if preferUpstream {
		names = []string{"upstream", "origin"}
	}
	for _, name := range names {
		u := GetRemoteUrl(cfg, name)
		if u != "" {
			return u, nil
		}
	}
	return "", nil
}

// GetRemoteUrl returns the remote URL from the given git config
func GetRemoteUrl(config *gitcfg.Config, name string) string {
	if config.Remotes != nil {
		return firstRemoteUrl(config.Remotes[name])
	}
	return ""
}

func firstRemoteUrl(remote *gitcfg.RemoteConfig) string {
	if remote != nil {
		urls := remote.URLs
		if urls != nil && len(urls) > 0 {
			return urls[0]
		}
	}
	return ""
}

func parseGitConfig(gitConf string) (*gitcfg.Config, error) {
	if gitConf == "" {
		return nil, fmt.Errorf("no GitConfDir defined")
	}
	cfg := gitcfg.NewConfig()
	data, err := os.ReadFile(gitConf)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s due to %s", gitConf, err)
	}

	err = cfg.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s due to %s", gitConf, err)
	}
	return cfg, nil
}
