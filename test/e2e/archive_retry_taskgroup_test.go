//go:build functional

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/argoproj/argo-workflows/v4/test/e2e/fixtures"
)

func (s *RetryTestSuite) TestArchiveRetryWithTaskGroupParameters() {
	s.Given().
		Workflow("@testdata/archive-retry-taskgroup-bug.yaml").
		When().
		SubmitWorkflow().
		WaitForWorkflow(fixtures.ToBeFailed).
		Then().
		ExpectWorkflow(func(t *testing.T, metadata *metav1.ObjectMeta, status *v1alpha1.WorkflowStatus) {
			assert.Equal(t, v1alpha1.WorkflowPhase("Failed"), status.Phase)
		})
	
	// Archive the workflow
	uid := s.GetWorkflowUID()
	s.e().POST("/api/v1/archived-workflows").
		WithQuery("listOptions.fieldSelector", "metadata.namespace=argo,metadata.name=archive-retry-taskgroup-").
		Then().
		Expect().
		Status(200).
		JSON().
		Path("$.items[0].metadata.uid").
		NotNull()
	
	// Now retry with nodeFieldSelector and parameters
	// This should NOT delete the successful frame-processing TaskGroup children,
	// but according to the bug, it does.
	s.Given().
		Workflow("@testdata/archive-retry-taskgroup-bug.yaml").
		When().
		SubmitWorkflow().
		WaitForWorkflow(fixtures.ToBeFailed).
		Then().
		ExpectWorkflow(func(t *testing.T, metadata *metav1.ObjectMeta, status *v1alpha1.WorkflowStatus) {
			// Find the UID of the archived workflow
			uid := s.GetWorkflowUID()
			
			// Retry with nodeFieldSelector pointing to post-processing node
			// and parameters to override image
			s.e().PUT("/api/v1/archived-workflows/"+uid+"/retry").
				WithBytes([]byte(`{
					"nodeFieldSelector": "displayName=post-processing",
					"restartSuccessful": false,
					"parameters": [
						"cpu_image=argoproj/argosay:v3"
					]
				}`)).
				Expect().
				Status(200).
				JSON().
				Path("$.metadata.name").
				NotNull()
		})
}
