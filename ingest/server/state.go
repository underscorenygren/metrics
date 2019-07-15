package server

import (
	"github.com/underscorenygren/metrics/producer"
	"io/ioutil"
	"net/http"
)

func (s *state) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		failed := producer.PutRecord(s.p, body)
		if failed == nil || len(failed) == 0 {
			s.logger.Debug("successfully put record")
		} else {
			s.logger.Error("failed to put logging record")
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
