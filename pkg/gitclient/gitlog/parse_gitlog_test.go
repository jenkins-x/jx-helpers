package gitlog_test

import (
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/gitlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	gitLogOutput = `
commit 1eecedcda6c17b38b2c2b055f45c9587d3befccc
Author: James Strachan <james.strachan@gmail.com>
Date:   Mon Dec 14 17:36:08 2020 +0000

    Update rebase.yaml

commit 617b6639284c4934f867833d41c4efe78d4f7e80
Author: James Strachan <james.strachan@gmail.com>
Date:   Mon Dec 14 17:33:32 2020 +0000

    Update rebase.yaml

commit a436c3808d03f731d3648e6836bdb32544c27498
Author: jenkins-x-bot <jenkins-x@googlegroups.com>
Date:   Mon Dec 14 16:59:34 2020 +0000

    chore: regenerated

    /pipeline cancel

commit de3eab72b9b1876e30e6dfc3f8ada3c76acaef45
Author: jstrachan <jenkins-x@googlegroups.com>
Date:   Mon Dec 14 16:58:25 2020 +0000

    chore: promote nodey534 to version 1.0.10 in Staging environment`
)

func TestParseGitLog(t *testing.T) {
	commits := gitlog.ParseGitLog(gitLogOutput)

	require.Len(t, commits, 4, "number of commits")

	assert.Equal(t, "1eecedcda6c17b38b2c2b055f45c9587d3befccc", commits[0].SHA, "commit[0].SHA")
	assert.Equal(t, "James Strachan <james.strachan@gmail.com>", commits[0].Author, "commit[0].Author")
	assert.Equal(t, "Mon Dec 14 17:36:08 2020 +0000", commits[0].Date, "commit[0].Date")

	assert.Equal(t, "617b6639284c4934f867833d41c4efe78d4f7e80", commits[1].SHA, "commit[1].SHA")
	assert.Equal(t, "James Strachan <james.strachan@gmail.com>", commits[1].Author, "commit[1].Author")

	assert.Equal(t, "a436c3808d03f731d3648e6836bdb32544c27498", commits[2].SHA, "commit[2].SHA")
	assert.Equal(t, "jenkins-x-bot <jenkins-x@googlegroups.com>", commits[2].Author, "commit[2].Author")
	assert.Equal(t, "chore: regenerated\n\n/pipeline cancel\n", commits[2].Comment, "commit[2].Comment")

	assert.Equal(t, "de3eab72b9b1876e30e6dfc3f8ada3c76acaef45", commits[3].SHA, "commit[3].SHA")
	assert.Equal(t, "jstrachan <jenkins-x@googlegroups.com>", commits[3].Author, "commit[3].Author")
	assert.Equal(t, "chore: promote nodey534 to version 1.0.10 in Staging environment", commits[3].Comment, "commit[3].Comment")
}
