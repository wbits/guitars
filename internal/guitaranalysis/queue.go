package guitaranalysis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// JobQueue enqueues async analysis jobs.
type JobQueue interface {
	Enqueue(ctx context.Context, job Job) error
}

// SQSQueue publishes jobs to an SQS queue.
type SQSQueue struct {
	Client   *sqs.Client
	QueueURL string
}

func (q *SQSQueue) Enqueue(ctx context.Context, job Job) error {
	if q == nil || q.Client == nil || strings.TrimSpace(q.QueueURL) == "" {
		return fmt.Errorf("analysis queue is not configured")
	}
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}
	_, err = q.Client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(q.QueueURL),
		MessageBody: aws.String(string(body)),
	})
	return err
}

// MemoryQueue stores jobs in memory for tests.
type MemoryQueue struct {
	mu    sync.Mutex
	Jobs  []Job
	Err   error
	OnSend func(job Job)
}

func (q *MemoryQueue) Enqueue(_ context.Context, job Job) error {
	if q == nil {
		return fmt.Errorf("analysis queue is not configured")
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.Err != nil {
		return q.Err
	}
	q.Jobs = append(q.Jobs, job)
	if q.OnSend != nil {
		q.OnSend(job)
	}
	return nil
}

func (q *MemoryQueue) Len() int {
	if q == nil {
		return 0
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.Jobs)
}
