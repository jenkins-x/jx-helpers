package scmhelpers_test

import (
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-helpers/v3/pkg/scmhelpers"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResponseNotFoundShouldBeFalseWhenNoResponseProvided(t *testing.T) {
	f := scmhelpers.IsScmResponseNotFound(nil)

	assert.Falsef(t, f, "should be false when response is nil")
}

func TestResponseNotFoundShouldBeFalseWhenResponseStatusIsOk(t *testing.T) {
	res := new(scm.Response)
	res.Status = 200

	f := scmhelpers.IsScmResponseNotFound(res)

	assert.Falsef(t, f, "should be false when response has 200 status")
}

func TestResponseNotFoundShouldBeTrueWhenResponseStatusIsNotFound(t *testing.T) {
	res := new(scm.Response)
	res.Status = 404

	f := scmhelpers.IsScmResponseNotFound(res)

	assert.Truef(t, f, "should be true when response has 404 status")
}
