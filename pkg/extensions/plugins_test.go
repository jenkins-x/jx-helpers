// +build unit

package extensions_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/extensions"
	"github.com/jenkins-x/jx-helpers/v3/pkg/testhelpers"

	"github.com/jenkins-x/jx-logging/v3/pkg/log"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	jenkinsv1 "github.com/jenkins-x/jx-api/pkg/apis/jenkins.io/v1"

	"github.com/jenkins-x/jx-api/pkg/util"
	"github.com/stretchr/testify/assert"
)

const (
	binDirNs   = "jx-test"
	port       = "0"
	name       = "jx-test.test-plugin"
	version    = "0.0.1"
	testString = "Testing123"
)

func TestEnsurePluginInstalled(t *testing.T) {
	testhelpers.SkipForWindows(t, "plugins do not work on windows - and this test will always fail."+""+
		"it is a valid failure - but holds up windows development.  See https://github.com/jenkins-x/jx/issues/2677")

	// Remove any existing cruft
	testPluginBinDir, err := ioutil.TempDir("", "")
	assert.NoError(t, err, "Error creating temp dir")
	t.Logf("using the test dir %s", testPluginBinDir)

	srv, port := serveTestScript(t)
	defer func() {
		err = srv.Close()
		assert.NoError(t, err, "Error getting plugin bin dir for namespace jx-test")
	}()
	testPlugin := jenkinsv1.Plugin{
		ObjectMeta: v1.ObjectMeta{
			Namespace: binDirNs,
		},
		Spec: jenkinsv1.PluginSpec{
			Description: "Test Plugin",
			Binaries: []jenkinsv1.Binary{
				{
					URL:    fmt.Sprintf("http://%s:%d/jx-test", "localhost", port),
					Goarch: "amd64",
					Goos:   "Windows",
				},
				{
					URL:    fmt.Sprintf("http://%s:%d/jx-test", "localhost", port),
					Goarch: "amd64",
					Goos:   "Darwin",
				},
				{
					URL:    fmt.Sprintf("http://%s:%d/jx-test", "localhost", port),
					Goarch: "amd64",
					Goos:   "Linux",
				},
			},
			Version:    version,
			Name:       name,
			SubCommand: "test-plugin",
			Group:      "Test Plugins",
		},
	}
	path, err := extensions.EnsurePluginInstalled(testPlugin, testPluginBinDir)
	assert.NoError(t, err, "Error ensuring plugin is installed")
	assert.EqualValues(t, filepath.Join(testPluginBinDir, fmt.Sprintf("%s-%s", name, version)), path, "Actual path is not equal to expected path")
	cmd := util.Command{
		Name: path,
	}
	res, err := cmd.RunWithoutRetry()
	assert.NoError(t, err, "Error running plugin")
	assert.EqualValues(t, testString, res)
}

func serveTestScript(t *testing.T) (*http.Server, int) {

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", "0.0.0.0", port))
	if err != nil {
		panic(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	srv := &http.Server{}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "#!/bin/sh\necho %s\n", testString)
	})
	go func() {
		if err := srv.Serve(listener); err != nil && err.Error() != "http: Server closed" {
			log.Logger().Errorf("Error starting HTTP server \n%v", err)
			assert.NoError(t, err, "Error starting HTTP server to serve test plugin script")
		}
	}()
	return srv, port
}
