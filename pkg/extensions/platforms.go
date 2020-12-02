package extensions

import jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"

// Platform represents a platform for binaries
type Platform struct {
	Goarch string
	Goos   string
}

var (
	// DefaultPlatforms the default list of platforms to create plugins for
	DefaultPlatforms = []Platform{
		{
			Goarch: "amd64",
			Goos:   "Windows",
		},
		{
			Goarch: "amd64",
			Goos:   "Darwin",
		},
		{
			Goarch: "amd64",
			Goos:   "Linux",
		},
		{
			Goarch: "arm",
			Goos:   "Linux",
		},
		{
			Goarch: "386",
			Goos:   "Linux",
		},
	}
)

// Extension returns the default distribution extension; `tar.gz` or `zip` for windows
func (p Platform) Extension() string {
	if p.IsWindows() {
		return "zip"
	}
	return "tar.gz"
}

// IsWindows returns true if the platform is windows
func (p Platform) IsWindows() bool {
	return p.Goos == "Windows"
}

// CreateBinaries a helper function to create the binary resources for the platforms for a given callback
func CreateBinaries(createURLFn func(Platform) string) []jenkinsv1.Binary {
	var answer []jenkinsv1.Binary
	for _, p := range DefaultPlatforms {
		u := createURLFn(p)
		if u != "" {
			answer = append(answer, jenkinsv1.Binary{
				Goarch: p.Goarch,
				Goos:   p.Goos,
				URL:    u,
			})
		}
	}
	return answer
}
