package waithelpers

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"
	batchv1client "k8s.io/client-go/kubernetes/typed/batch/v1"
)

func WaitForCronJobState(
	ctx context.Context,
	client batchv1client.CronJobInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*batchv1.CronJob) (bool, error),
) (*batchv1.CronJob, error) {
	return WaitForObjectState[*batchv1.CronJob, *batchv1.CronJobList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForCronJobDeletion(
	ctx context.Context,
	client batchv1client.CronJobInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*batchv1.CronJob, *batchv1.CronJobList](
		ctx,
		client,
		nn,
		uid,
	)
}

func WaitForJobState(
	ctx context.Context,
	client batchv1client.JobInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*batchv1.Job) (bool, error),
) (*batchv1.Job, error) {
	return WaitForObjectState[*batchv1.Job, *batchv1.JobList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForJobDeletion(
	ctx context.Context,
	client batchv1client.JobInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*batchv1.Job, *batchv1.JobList](
		ctx,
		client,
		nn,
		uid,
	)
}
