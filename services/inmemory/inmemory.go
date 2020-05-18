package inmemory

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/adzeitor/background"
	"github.com/google/uuid"
)

// Service represents in-memory service for store information about background
// jobs.
type Service struct {
	jobs map[string]background.Job
	sync.Mutex
}

// New creates new in-memory jobs service.
func New() *Service {
	jobs := make(map[string]background.Job)
	return &Service{
		jobs: jobs,
	}
}

// JobStarted saves initial information about new background job.
func (s *Service) JobStarted(
	_ context.Context,
	kind string,
) (background.Job, error) {
	s.Lock()
	defer s.Unlock()

	now := time.Now()
	job := background.Job{
		ID:        uuid.New().String(),
		Status:    background.StatusWorking,
		Kind:      kind,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.jobs[job.ID] = job
	return job, nil
}

// JobCompleted saves that background job is completed.
func (s *Service) JobCompleted(
	_ context.Context,
	id string,
	response background.Response,
) error {
	s.Lock()
	defer s.Unlock()

	job := s.jobs[id]
	job.Status = background.StatusCompleted
	job.Response = &response
	job.UpdatedAt = time.Now()
	s.jobs[id] = job
	return nil
}

// Get retrieves job information by job ID.
func (s *Service) Get(_ context.Context, id string) (background.Job, error) {
	s.Lock()
	defer s.Unlock()

	job, ok := s.jobs[id]
	if !ok {
		return background.Job{}, errors.New("not found")
	}
	return job, nil
}

// Jobs retrieves all jobs.
func (s *Service) Jobs() ([]background.Job, error) {
	s.Lock()
	defer s.Unlock()

	var jobs []background.Job
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.Before(jobs[j].CreatedAt)
	})
	return jobs, nil
}

// Jobs sets that job is working.
func (s *Service) Ping(_ context.Context, id string) error {
	s.Lock()
	defer s.Unlock()

	job := s.jobs[id]
	if job.Status == background.StatusCompleted {
		return nil
	}

	job.Status = background.StatusWorking
	job.UpdatedAt = time.Now()
	s.jobs[id] = job
	return nil
}
