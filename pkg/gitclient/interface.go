package gitclient

// Interface a simple interface to performing git commands
// which is easy to fake for testing
type Interface interface {
	// Command command runs the git sub command such as 'commit' or 'clone'
	// in the given directory with the arguments
	Command(dir string, args ...string) (string, error)
}
