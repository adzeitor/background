package background

import (
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
