package background

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"
)

// Background converts handlers to background handlers.
type Background struct {
	Service JobService
	Logger  Logger

	executor func(func() error)
}

// JobService manages background jobs information.
type JobService interface {
	JobStarted(ctx context.Context, kind string) (Job, error)
	JobCompleted(ctx context.Context, id string, response Response) error
	Ping(ctx context.Context, id string) error
}

// Logger manages logging.
type Logger interface {
	Println(...interface{})
}

// NewMiddleware creates new middleware that use service as backend for
// job information updating.
// Currently all handlers executed in goroutines.
func NewMiddleware(service JobService, logger Logger) *Background {
	bg := &Background{
		Service: service,
		Logger:  logger,
	}
	bg.executor = bg.goroutineExecutor
	return bg
}

// InBackground converts handler to background handler.
// It respond immediately with job ID that can be used to track
// status and getting response.
func (bg *Background) InBackground(
	origHandler http.Handler,
	kind string,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		job, err := bg.Service.JobStarted(r.Context(), kind)
		if err != nil {
			bg.Logger.Println(err)
			http.Error(w, "cannot start background job", http.StatusInternalServerError)
			return
		}

		clonedRequest, err := cloneRequest(r)
		if err != nil {
			http.Error(w, "failed to get request body", http.StatusInternalServerError)
			return
		}
		bg.executor(func() error {
			return bg.serve(job, clonedRequest, origHandler)
		})

		err = jobStartedResponse(w, job)
		if err != nil {
			bg.Logger.Println(err)
		}
	})
}

func (bg *Background) goroutineExecutor(handler func() error) {
	go func() {
		err := handler()
		if err != nil {
			bg.Logger.Println(err)
		}
	}()
}

func (bg *Background) serve(
	job Job,
	r *http.Request,
	origHandler http.Handler,
) error {
	// FIXME: It should be r.Context() instead of new context.Background, but
	// original context probably will be closed, because serve of original
	// request is completed. Which can lead to immediately closing of superviseJob.
	// Maybe implementing ContextWithoutDone will fix this issue.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go bg.superviseJob(ctx, job)

	recorder := httptest.NewRecorder()
	origHandler.ServeHTTP(recorder, r)

	result := recorder.Result()
	response, err := newResponse(result)
	if err != nil {
		// FIXME: default parameters is no good...
		// Maybe JobFailed method?
		return bg.Service.JobCompleted(ctx, job.ID, Response{})
	}
	return bg.Service.JobCompleted(ctx, job.ID, response)
}

func (bg *Background) superviseJob(ctx context.Context, job Job) {
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			err := bg.Service.Ping(ctx, job.ID)
			if err != nil {
				bg.Logger.Println("ping error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func jobStartedResponse(w http.ResponseWriter, job Job) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)

	response := struct {
		ID string `json:"id"`
	}{
		ID: job.ID,
	}
	return json.NewEncoder(w).Encode(response)
}

func cloneRequest(r *http.Request) (*http.Request, error) {
	clonedRequest := r.Clone(WithNoCancelContext(r.Context()))
	if r.Body != nil {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		clonedRequest.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}

	return clonedRequest, nil
}

func newResponse(response *http.Response) (Response, error) {
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return Response{}, err
	}
	return Response{
		StatusCode: response.StatusCode,
		Header:     response.Header,
		Body:       string(body),
	}, nil
}
