package waithelpers

import (
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestWaitForCronJobState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForCronJobState(t *testing.T) {
	t.Parallel()

	cronJob := &batchv1.CronJob{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(cronJob).BatchV1().CronJobs(cronJob.GetNamespace()),
		cronJob.GetName(),
		WaitForCronJobState,
	)
}

// TestWaitForCronJobDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForCronJobDeletion(t *testing.T) {
	t.Parallel()

	cronJob := &batchv1.CronJob{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"CronJob",
		cronJob,
		fake.NewClientset(cronJob).BatchV1().CronJobs(testNamespaceName),
		deleteObject,
		WaitForCronJobDeletion,
	)
}

// TestWaitForJobState tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectState tests the specific cases.
func TestWaitForJobState(t *testing.T) {
	t.Parallel()

	job := &batchv1.Job{ObjectMeta: newTestMeta("any")}
	testWaitForState(
		t,
		fake.NewClientset(job).BatchV1().Jobs(job.GetNamespace()),
		job.GetName(),
		WaitForJobState,
	)
}

// TestWaitForJobDeletion tests the minimal possible scenario just to be sure the function can
// be instantiated and the parameters are passed through.
// TestWaitForObjectDeletion tests the specific cases.
func TestWaitForJobDeletion(t *testing.T) {
	t.Parallel()

	job := &batchv1.Job{ObjectMeta: newTestMeta("")}
	testWaitForDeletion(
		t,
		"Job",
		job,
		fake.NewClientset(job).BatchV1().Jobs(testNamespaceName),
		deleteObject,
		WaitForJobDeletion,
	)
}
