package server

import (
	"github.com/underscorenygren/metrics/producer"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

func (s *server) handleError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	s.logger.Error("error on request", zap.Error(err))
}

func (s *server) isHealthcheckRequest(req *http.Request) bool {
	return s.healthcheck != nil &&
		(req.URL != nil &&
			req.URL.Path == s.healthcheck.Path) &&
		req.Method == s.healthcheck.Method
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		s.handleError(err, w)
		return
	}
	s.logger.Debug("received request", zap.ByteString("body", body))

	if s.isHealthcheckRequest(req) {
		s.logger.Debug("healtcheck request")
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx := req.Context()
	if s.contextMaker != nil {
		if ctx, err = s.contextMaker(req); err != nil {
			s.handleError(err, w)
			return
		}
	}
	if s.middleware != nil {
		s.logger.Debug("executing middleware")
		if body, err = s.middleware.Transform(ctx, body); err != nil {
			s.handleError(err, w)
			return
		}
		s.logger.Debug("middleware succeeded", zap.ByteString("body", body))
	}
	failed := producer.PutRecord(s.p, body)
	if failed == nil || len(failed) == 0 {
		s.logger.Debug("successfully put record")
	} else {
		s.logger.Error("failed to put logging record", zap.ByteString("failed", failed))
	}
	w.WriteHeader(http.StatusNoContent)
}
