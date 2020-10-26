package giturl

import (
	"fmt"
	"io"
	"strings"

	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
)

// PrintCreateRepositoryGenerateAccessToken prints the access token URL of a Git repository
func PrintCreateRepositoryGenerateAccessToken(kind string, serverURL string, username string, o io.Writer) {
	tokenUrl := ProviderAccessTokenURL(kind, serverURL, username)

	fmt.Fprintf(o, "To work with git provider %s we need an API Token\n", serverURL)
	fmt.Fprintf(o, "Please click this URL and generate a token \n%s\n\n", termcolor.ColorInfo(tokenUrl))
	fmt.Fprint(o, "Then COPY the token and enter it below:\n\n")
}

func ProviderAccessTokenURL(kind string, url string, username string) string {
	switch kind {
	case KindBitBucketCloud:
		// TODO pass in the username
		return BitBucketCloudAccessTokenURL(url, username)
	case KindBitBucketServer:
		return BitBucketServerAccessTokenURL(url)
	case KindGitea:
		return GiteaAccessTokenURL(url)
	case KindGitlab:
		return GitlabAccessTokenURL(url)
	default:
		return GitHubAccessTokenURL(url)
	}
}

func BitBucketCloudAccessTokenURL(url string, username string) string {
	// TODO with github we can default the scopes/flags we need on a token via adding
	// ?scopes=repo,read:user,user:email,write:repo_hook
	//
	// is there a way to do that for bitbucket?
	return stringhelpers.UrlJoin(url, "/account/user", username, "/app-passwords/new")
}

// BitBucketServerAccessTokenURL generates the access token URL
func BitBucketServerAccessTokenURL(url string) string {
	// TODO with github we can default the scopes/flags we need on a token via adding
	// ?scopes=repo,read:user,user:email,write:repo_hook
	//
	// is there a way to do that for bitbucket?
	return stringhelpers.UrlJoin(url, "/plugins/servlet/access-tokens/manage")
}

// GiteaAccessTokenURL returns the URL to generate an access token
func GiteaAccessTokenURL(url string) string {
	return stringhelpers.UrlJoin(url, "/user/settings/applications")
}

// GitlabAccessTokenURL returns the URL to click on to generate a personal access token for the Git provider
func GitlabAccessTokenURL(url string) string {
	return stringhelpers.UrlJoin(url, "/profile/personal_access_tokens")
}

func GitHubAccessTokenURL(url string) string {
	if strings.Index(url, "://") < 0 {
		url = "https://" + url
	}
	return stringhelpers.UrlJoin(url, "/settings/tokens/new?scopes=repo,read:user,read:org,user:email,admin:repo_hook,delete_repo,write:packages,read:packages,write:discussion,workflow")

}
