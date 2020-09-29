package scmhelpers_test

import (
	"testing"

	jxfake "github.com/jenkins-x/jx-api/pkg/client/clientset/versioned/fake"
	"github.com/jenkins-x/jx-helpers/pkg/kube/jxenv"
	"github.com/jenkins-x/jx-helpers/pkg/scmhelpers"
	"github.com/jenkins-x/jx-helpers/pkg/stringhelpers"
	"github.com/jenkins-x/jx-helpers/pkg/testhelpers/testjx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverGitKind(t *testing.T) {
	ns := "jx"
	devEnv := jxenv.CreateDefaultDevEnvironment(ns)
	devEnv.Namespace = ns
	devEnv.Spec.Source.URL = "https://github.com/myorg/myrepo.git"

	owner := "myorg"
	repo := "myrepo"
	provider := "https://mgitlab.com"
	kind := "gitlab"
	sr := testjx.CreateSourceRepository(ns, owner, repo, kind, provider)

	o := &scmhelpers.Options{}

	o.Namespace = ns
	o.JXClient = jxfake.NewSimpleClientset(sr)
	o.SourceURL = stringhelpers.UrlJoin(provider, owner, repo)
	o.Owner = owner
	o.Repository = repo
	o.Branch = "master"
	o.GitToken = "dummytoken"

	err := o.Validate()
	if err != nil {
		require.NoError(t, err, "failed to validate")
	}

	t.Logf("found git kind %s\n", o.GitKind)
	assert.Equal(t, kind, o.GitKind, "should have discovered the GitKind")
}
