package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// networkOperationPollInterval is how long to wait between polls of a queued network deploy or
// teardown. It is a variable so tests can shorten it.
var networkOperationPollInterval = 5 * time.Second

// networkOperationAPIModel is a queued deploy or undeploy job (NetworkOperation in the v3 API).
// Network is filled in once the deploy has written the Network row.
type networkOperationAPIModel struct {
	JobID   int64              `json:"job_id"`
	Jobname string             `json:"jobname"`
	Status  jobNestedString    `json:"status"`
	Result  jobNestedString    `json:"result"`
	Message *string            `json:"message"`
	Network *networkV3APIModel `json:"network"`
}

type networkOperationState int

const (
	networkOperationPending networkOperationState = iota
	networkOperationSucceeded
	networkOperationFailed
)

// MCS jobs report progress in `status` (e.g. RUNNING, then PROCESSED once the job has stopped)
// and the outcome in `result` (e.g. PENDING, then SUCCESS or EXCEPTION). Values are matched
// case-insensitively. The result decides the outcome; a job whose status says it has stopped
// without a success result counts as failed, so an unknown failure result cannot leave Terraform
// waiting until the timeout.
var (
	networkOperationSuccessResults = map[string]bool{
		"success": true, "successful": true, "succeeded": true,
	}
	networkOperationFailureResults = map[string]bool{
		"exception": true, "failed": true, "failure": true, "error": true,
		"cancelled": true, "canceled": true, "aborted": true,
	}
	networkOperationStoppedStatuses = map[string]bool{
		"processed": true, "finished": true, "completed": true, "done": true,
		"failed": true, "error": true, "cancelled": true, "canceled": true, "aborted": true,
	}
)

func normalizedJobValue(v jobNestedString) string {
	if v.Value == nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(*v.Value))
}

func (op *networkOperationAPIModel) state() networkOperationState {
	status, result := normalizedJobValue(op.Status), normalizedJobValue(op.Result)
	switch {
	case networkOperationSuccessResults[result]:
		return networkOperationSucceeded
	case networkOperationFailureResults[result], networkOperationStoppedStatuses[status]:
		return networkOperationFailed
	default:
		return networkOperationPending
	}
}

// describe summarises the job for error messages.
func (op *networkOperationAPIModel) describe() string {
	s := fmt.Sprintf("job %d", op.JobID)
	if op.Jobname != "" {
		s += fmt.Sprintf(" (%s)", op.Jobname)
	}
	if v := op.Status.Value; v != nil {
		s += fmt.Sprintf(", status %q", *v)
	}
	if v := op.Result.Value; v != nil {
		s += fmt.Sprintf(", result %q", *v)
	}
	if op.Message != nil && *op.Message != "" {
		s += fmt.Sprintf(": %s", *op.Message)
	}
	return s
}

// waitForNetworkOperation polls a networking job until it succeeds, fails, or ctx is done
// (callers bound ctx with the resource timeout). It always returns the most recent job it saw,
// so a caller can still pick up a Network the job created before failing or timing out.
func waitForNetworkOperation(ctx context.Context, c *apiclient.Client, op *networkOperationAPIModel) (*networkOperationAPIModel, error) {
	last := op
	for {
		switch last.state() {
		case networkOperationSucceeded:
			return last, nil
		case networkOperationFailed:
			return last, fmt.Errorf("network operation failed: %s", last.describe())
		}

		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return last, fmt.Errorf("timed out waiting for network operation: %s", last.describe())
			}
			return last, ctx.Err()
		case <-time.After(networkOperationPollInterval):
		}

		var next networkOperationAPIModel
		if err := c.Get(ctx, fmt.Sprintf("/api/v3/networking/operations/%d/", last.JobID), &next); err != nil {
			return last, fmt.Errorf("polling network operation %d: %w", last.JobID, err)
		}
		if next.JobID == 0 {
			next.JobID = last.JobID
		}
		if next.Network == nil {
			next.Network = last.Network
		}
		tflog.Debug(ctx, "polled network operation", map[string]interface{}{
			"job_id": next.JobID,
			"status": normalizedJobValue(next.Status),
			"result": normalizedJobValue(next.Result),
		})
		last = &next
	}
}

// waitForNetworkGone polls a network until the API answers 404, or ctx is done.
func waitForNetworkGone(ctx context.Context, c *apiclient.Client, networkPath string) error {
	for {
		var n networkV3APIModel
		err := c.Get(ctx, networkPath, &n)
		if apiclient.IsNotFound(err) {
			return nil
		}
		if err != nil && ctx.Err() == nil {
			return fmt.Errorf("checking whether network was removed: %w", err)
		}

		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("timed out waiting for network %s to be removed", networkPath)
			}
			return ctx.Err()
		case <-time.After(networkOperationPollInterval):
		}
	}
}
