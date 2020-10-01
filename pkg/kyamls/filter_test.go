package kyamls_test

import (
	"path/filepath"
	"testing"

	"github.com/jenkins-x/jx-helpers/v3/pkg/kyamls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestFilter(t *testing.T) {
	testCases := []struct {
		filter   kyamls.Filter
		file     string
		expected bool
	}{
		{
			filter: kyamls.Filter{
				Selector: map[string]string{
					"chart":                        "myapp-0.0.12",
					"gitops.jenkins-x.io/pipeline": "environment",
				},
			},
			file:     "deployment.yaml",
			expected: true,
		},
		{
			filter: kyamls.Filter{
				Kinds: []string{"Service"},
				Selector: map[string]string{
					"beer": "stella",
				},
			},
			file:     "service.yaml",
			expected: true,
		},
		{
			filter: kyamls.Filter{
				Kinds: []string{"Service"},
				Selector: map[string]string{
					"beer": "veltins",
				},
			},
			file:     "service.yaml",
			expected: false,
		},
		{
			filter: kyamls.Filter{
				Selector: map[string]string{
					"beer": "stella",
				},
			},
			file:     "service.yaml",
			expected: true,
		},
		{
			filter: kyamls.Filter{
				Selector: map[string]string{
					"beer": "stella",
					"wine": "shiraz",
				},
			},
			file:     "service.yaml",
			expected: true,
		},
		{
			filter: kyamls.Filter{
				Selector: map[string]string{
					"beer":   "stella",
					"cheese": "edam",
					"wine":   "shiraz",
				},
			},
			file:     "service.yaml",
			expected: true,
		},
		{
			filter: kyamls.Filter{
				Selector: map[string]string{
					"beer":   "stella",
					"cheese": "edam",
					"wine":   "malbec",
				},
			},
			file:     "service.yaml",
			expected: false,
		},
		{
			filter: kyamls.Filter{
				Selector: map[string]string{
					"wine": "malbec",
				},
			},
			file:     "service.yaml",
			expected: false,
		},
		{
			filter: kyamls.Filter{
				Names: []string{"myapp-myapp"},
			},
			file:     "deployment.yaml",
			expected: true,
		},
		{
			filter: kyamls.Filter{
				Kinds: []string{"apps/"},
			},
			file:     "deployment.yaml",
			expected: true,
		},
		{
			filter: kyamls.Filter{
				Kinds: []string{"apps/v1/"},
			},
			file:     "deployment.yaml",
			expected: true,
		},
		{
			filter: kyamls.Filter{
				Kinds: []string{"apps/v2/"},
			},
			file:     "deployment.yaml",
			expected: false,
		},
		{
			filter: kyamls.Filter{
				Kinds: []string{"apps/Deployment"},
			},
			file:     "deployment.yaml",
			expected: true,
		},
		{
			filter: kyamls.Filter{
				Kinds: []string{"apps/v1/Deployment"},
			},
			file:     "deployment.yaml",
			expected: true,
		},
		{
			filter: kyamls.Filter{
				Kinds: []string{"v1/Service"},
			},
			file:     "service.yaml",
			expected: true,
		},
		{
			filter: kyamls.Filter{
				Kinds: []string{"v2/Service"},
			},
			file:     "service.yaml",
			expected: false,
		},
		{
			filter: kyamls.Filter{
				Kinds: []string{"Service"},
			},
			file:     "service.yaml",
			expected: true,
		},
		{
			filter: kyamls.Filter{
				KindsIgnore: []string{"ConfigMap"},
			},
			file:     "service.yaml",
			expected: true,
		},
		{
			filter: kyamls.Filter{
				Kinds: []string{"ConfigMap"},
			},
			file:     "service.yaml",
			expected: false,
		},
	}

	for _, tc := range testCases {
		file := filepath.Join("test_data", tc.file)
		filter := tc.filter
		node, err := yaml.ReadFile(file)
		require.NoError(t, err, "reading file %s", file)

		fn, err := filter.ToFilterFn()
		require.NoError(t, err, "creating filter function for file %s", file)
		require.NotNil(t, fn, "creating filter function for file %s filter %#v", file, filter)

		flag, err := fn(node, file)
		require.NoError(t, err, "evaluating filter function for file %s filter %#v", file, filter)

		assert.Equal(t, tc.expected, flag, "evaluating filter function for file %s for %#v", file, filter)

		t.Logf("evaluated filter %#v on file %s and got %#v\n", filter, file, flag)
	}
}
