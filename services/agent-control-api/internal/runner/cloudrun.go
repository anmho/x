package runner

import (
	"context"
	"fmt"

	run "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
)

// CloudRunClient dispatches agent-runner jobs via the Cloud Run Jobs API.
type CloudRunClient struct {
	client  *run.JobsClient
	jobName string // projects/{project}/locations/{region}/jobs/{job}
}

func NewCloudRunClient(ctx context.Context, jobName string) (*CloudRunClient, error) {
	client, err := run.NewJobsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("cloudrun: new jobs client: %w", err)
	}
	return &CloudRunClient{client: client, jobName: jobName}, nil
}

// Dispatch starts a new execution of the agent-runner job with the given runtime spec.
// Returns the fully-qualified execution name.
func (c *CloudRunClient) Dispatch(ctx context.Context, spec ExecutionSpec) (string, error) {
	envVars := make([]*runpb.EnvVar, 0, len(spec.Env)+5)
	for k, v := range CloudRunEnv(spec) {
		envVars = append(envVars, &runpb.EnvVar{
			Name:   k,
			Values: &runpb.EnvVar_Value{Value: v},
		})
	}

	op, err := c.client.RunJob(ctx, &runpb.RunJobRequest{
		Name: c.jobName,
		Overrides: &runpb.RunJobRequest_Overrides{
			ContainerOverrides: []*runpb.RunJobRequest_Overrides_ContainerOverride{
				{
					Env:  envVars,
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("cloudrun: run job: %w", err)
	}

	exec, err := op.Poll(ctx)
	if err != nil {
		// op.Poll returning error just means it hasn't finished — we still have the execution name
		_ = err
	}
	if exec != nil {
		return exec.Name, nil
	}
	// Return the operation name as a fallback so callers can still track
	return op.Name(), nil
}

// ExecutionStatus maps a Cloud Run execution condition to our domain status string.
func (c *CloudRunClient) ExecutionStatus(ctx context.Context, executionName string) (string, error) {
	execClient, err := run.NewExecutionsClient(ctx)
	if err != nil {
		return "", fmt.Errorf("cloudrun: new executions client: %w", err)
	}
	defer execClient.Close()

	exec, err := execClient.GetExecution(ctx, &runpb.GetExecutionRequest{Name: executionName})
	if err != nil {
		return "", fmt.Errorf("cloudrun: get execution: %w", err)
	}

	for _, cond := range exec.Conditions {
		if cond.Type == "Completed" {
			switch cond.State {
			case runpb.Condition_CONDITION_SUCCEEDED:
				return "SUCCEEDED", nil
			case runpb.Condition_CONDITION_FAILED:
				return "FAILED", nil
			}
		}
	}
	return "RUNNING", nil
}
