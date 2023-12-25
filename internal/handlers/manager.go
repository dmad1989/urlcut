package handlers

import "net/http"

func Manage(resp http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		CutterHandler(resp, req)
	case http.MethodGet:
		Redirect(resp, req)
	default:
		http.Error(resp, "", http.StatusBadRequest)
		// resp.WriteHeader(http.StatusBadRequest)
	}
}
