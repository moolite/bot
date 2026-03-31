package chain

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestThen_EmptyChain(t *testing.T) {
	is := is.New(t)

	h := Chain{}.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

	is.Equal(w.Code, http.StatusOK)
	is.Equal(w.Body.String(), "ok")
}

func TestThen_SingleMiddleware(t *testing.T) {
	is := is.New(t)

	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom", "set")
			next.ServeHTTP(w, r)
		})
	}

	h := Chain{mw}.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

	is.Equal(w.Header().Get("X-Custom"), "set")
	is.Equal(w.Code, http.StatusOK)
}

func TestThen_MultipleMiddlewares_ExecutionOrder(t *testing.T) {
	is := is.New(t)

	var order []string
	record := func(name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, name)
			next.ServeHTTP(w, r)
			order = append(order, name)
		})
	}

	mw1 := func(next http.Handler) http.Handler { return record("mw1", next) }
	mw2 := func(next http.Handler) http.Handler { return record("mw2", next) }
	mw3 := func(next http.Handler) http.Handler { return record("mw3", next) }

	h := Chain{mw1, mw2, mw3}.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	}))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

	is.Equal(order, []string{"mw1", "mw2", "mw3", "handler", "mw3", "mw2", "mw1"})
}

func TestThenFunc(t *testing.T) {
	is := is.New(t)

	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Via", "middleware")
			next.ServeHTTP(w, r)
		})
	}

	h := Chain{mw}.ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

	is.Equal(w.Header().Get("X-Via"), "middleware")
	is.Equal(w.Code, http.StatusNoContent)
}
