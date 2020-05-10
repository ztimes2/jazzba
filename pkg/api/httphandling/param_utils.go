package httphandling

import (
	"net/http"
	"strconv"

	"github.com/ztimes2/jazzba/pkg/api/p8n"

	"github.com/go-chi/chi"
)

func readStringPathParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func readIntPathParam(r *http.Request, key string) (int, bool) {
	param, err := strconv.Atoi(chi.URLParam(r, key))
	if err != nil {
		return 0, false
	}
	return param, true
}

func readStringQueryParam(r *http.Request, key string) string {
	if r.Form == nil {
		r.ParseForm()
	}
	return r.Form.Get(key)
}

func readIntQueryParam(r *http.Request, key string) (int, bool) {
	param, err := strconv.Atoi(readStringQueryParam(r, key))
	if err != nil {
		return 0, false
	}
	return param, true
}

func readStringQueryParams(r *http.Request, key string) []string {
	if r.Form == nil {
		r.ParseForm()
	}
	return r.Form[key]
}

func readIntQueryParams(r *http.Request, key string) ([]int, bool) {
	var convertedParams []int

	for _, param := range readStringQueryParams(r, key) {
		convertedParam, err := strconv.Atoi(param)
		if err != nil {
			return nil, false
		}

		convertedParams = append(convertedParams, convertedParam)
	}

	return convertedParams, true
}

const (
	defaultPaginationLimit  = 10
	defaultPaginationOffset = 0
)

func readPaginationParam(r *http.Request) p8n.Page {
	limit, ok := readIntQueryParam(r, "limit")
	if !ok {
		limit = defaultPaginationLimit
	}

	offset, ok := readIntQueryParam(r, "offset")
	if !ok {
		offset = defaultPaginationOffset
	}

	return p8n.NewPage(limit, offset)
}
