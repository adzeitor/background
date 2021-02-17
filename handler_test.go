package background

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackground_InBackground(t *testing.T) {
	jobID := "6d1ed4de-a8cc-42b8-a743-6713a93626d0"
	wantBody := "job result: 0x3333"
	service := newServiceMock(t)
	service.jobStarted = func(kind string) (Job, error) {
		return Job{ID: jobID}, nil
	}
	service.jobCompleted = func(id string, response Response) error {
		assert.Equal(t, jobID, id)
		assert.Equal(t, wantBody, response.Body)
		return nil
	}
	service.ping = func(id string) error {
		assert.Equal(t, jobID, id)
		return nil
	}

	bg := NewMiddleware(service, &log.Logger{})
	bg.executor = func(handler func() error) {
		_ = handler()
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(wantBody))
	}

	backgroundHandler := bg.InBackground(http.HandlerFunc(handler), "testkind")

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	backgroundHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.JSONEq(t, `{"id":"6d1ed4de-a8cc-42b8-a743-6713a93626d0"}`, w.Body.String())
}

func TestBackground_goroutineExecutor(t *testing.T) {
	bg := NewMiddleware(nil, &log.Logger{})

	wg := sync.WaitGroup{}
	wg.Add(1)
	handler := func() error {
		wg.Done()
		return nil
	}

	bg.goroutineExecutor(handler)
	wg.Wait()
}

func Test_cloneRequest(t *testing.T) {
	t.Run("context should preserve values", func(t *testing.T) {
		key := "key"
		ctx := context.WithValue(context.Background(), key, 42)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
		assert.NoError(t, err)

		cloned, err := cloneRequest(req)
		assert.NoError(t, err)

		assert.Equal(t, 42, cloned.Context().Value(key))
	})

	t.Run("cancellation context should not affect cloned request", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
		assert.NoError(t, err)

		cloned, err := cloneRequest(req)
		assert.NoError(t, err)

		cancel()
		assert.NoError(t, cloned.Context().Err())
	})
}
