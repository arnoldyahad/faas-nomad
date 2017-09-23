package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexellis/faas/gateway/requests"
	"github.com/hashicorp/faas-nomad/nomad"
	"github.com/hashicorp/nomad/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupReplicationReader(functionName string) (http.HandlerFunc, *httptest.ResponseRecorder, *http.Request) {
	mockJob = &nomad.MockJob{}
	rr := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test/test_function", nil)
	r = r.WithContext(context.WithValue(r.Context(), FunctionNameCTXKey, functionName))

	h := MakeReplicationReader(mockJob)

	return h, rr, r
}

func TestMiddlewareReturns404WhenNotFound(t *testing.T) {
	functionName := "notFound"
	h, rr, r := setupReplicationReader(functionName)
	mockJob.On("Info", nomad.JobPrefix+functionName, mock.Anything).Return(nil, nil, nil)

	h(rr, r)

	mockJob.AssertCalled(t, "Info", nomad.JobPrefix+functionName, mock.Anything)
	assert.Equal(t, rr.Code, http.StatusNotFound)
}

func TestReplicationRReturnsFunctionWhenFound(t *testing.T) {
	functionName := "tester"
	jobName := nomad.JobPrefix + functionName

	h, rr, r := setupReplicationReader(functionName)
	mockJob.On("Info", jobName, mock.Anything).
		Return(&api.Job{ID: &jobName}, nil, nil)

	h(rr, r)

	f := &requests.Function{}
	err := json.NewDecoder(rr.Body).Decode(f)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, rr.Header().Get("Content-Type"), "application/json")
	assert.Equal(t, functionName, f.Name)
}