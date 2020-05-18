package background

import (
	"context"
	"testing"
)

// serviceMock represents mock of Service interface.
type serviceMock struct {
	jobStarted   func(kind string) (Job, error)
	jobCompleted func(id string, response Response) error
	ping         func(id string) error
}

// newServiceMock creates new instance of serviceMock.
func newServiceMock(t testing.TB) *serviceMock {
	m := serviceMock{}
	return &m
}

// JobStarted mocks method of Service.
func (m *serviceMock) JobStarted(ctx context.Context, kind string) (Job, error) {
	return m.jobStarted(kind)
}

// JobCompleted mocks method of Service.
func (m *serviceMock) JobCompleted(ctx context.Context, id string, response Response) error {
	return m.jobCompleted(id, response)
}

// Ping mocks method of Service.
func (m *serviceMock) Ping(ctx context.Context, id string) error {
	return m.ping(id)
}
