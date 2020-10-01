package activities_test

import (
	"testing"

	v1 "github.com/jenkins-x/jx-api/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/activities"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUpdateStatus(t *testing.T) {
	ns := "jx"

	testCases := []struct {
		activity *v1.PipelineActivity
		expected v1.ActivityStatusType
	}{
		{
			expected: v1.ActivityStatusTypePending,
			activity: &v1.PipelineActivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "all-pending",
					Namespace: ns,
				},
				Spec: v1.PipelineActivitySpec{
					Steps: []v1.PipelineActivityStep{
						{
							Kind: v1.ActivityStepKindTypeStage,
							Stage: &v1.StageActivityStep{
								CoreActivityStep: v1.CoreActivityStep{
									Name:   "step-1",
									Status: v1.ActivityStatusTypePending,
								},
							},
						},
						{
							Kind: v1.ActivityStepKindTypeStage,
							Stage: &v1.StageActivityStep{
								CoreActivityStep: v1.CoreActivityStep{
									Name:   "step-2",
									Status: v1.ActivityStatusTypePending,
								},
							},
						},
					},
				},
			},
		},
		{
			expected: v1.ActivityStatusTypeSucceeded,
			activity: &v1.PipelineActivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "all-succeeded",
					Namespace: ns,
				},
				Spec: v1.PipelineActivitySpec{
					Steps: []v1.PipelineActivityStep{
						{
							Kind: v1.ActivityStepKindTypeStage,
							Stage: &v1.StageActivityStep{
								Steps: []v1.CoreActivityStep{
									{
										Name:   "do-something",
										Status: v1.ActivityStatusTypeSucceeded,
									},
								},
							},
						},
						{
							Kind: v1.ActivityStepKindTypePromote,
							Promote: &v1.PromoteActivityStep{
								CoreActivityStep: v1.CoreActivityStep{
									Name:   "promote-to-staging",
									Status: v1.ActivityStatusTypeSucceeded,
								},
								Environment: "staging",
							},
						},
					},
				},
			},
		},
		{
			expected: v1.ActivityStatusTypeRunning,
			activity: &v1.PipelineActivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "promoted-but-step-running",
					Namespace: ns,
				},
				Spec: v1.PipelineActivitySpec{
					Steps: []v1.PipelineActivityStep{
						{
							Kind: v1.ActivityStepKindTypeStage,
							Stage: &v1.StageActivityStep{
								Steps: []v1.CoreActivityStep{
									{
										Name:   "do-something",
										Status: v1.ActivityStatusTypeRunning,
									},
								},
							},
							Promote: nil,
							Preview: nil,
						},
						{
							Kind:  v1.ActivityStepKindTypeStage,
							Stage: &v1.StageActivityStep{},
							Promote: &v1.PromoteActivityStep{
								CoreActivityStep: v1.CoreActivityStep{
									Name:   "promote-to-staging",
									Status: v1.ActivityStatusTypeSucceeded,
								},
								Environment: "staging",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		activities.UpdateStatus(tc.activity, false, nil)
		actual := tc.activity.Spec.Status
		assert.Equal(t, tc.expected, actual, "for PipelineActivity %s", tc.activity.Name)
	}
}
