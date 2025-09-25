package middleware

import (
	"fmt"
	"net/http"

	"github.com/mustafacaglarkara/webdev/pkg/logx"
	"github.com/mustafacaglarkara/webdev/pkg/web"
)

// Recover middleware: panik yakalar, loglar ve 500 döner.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// structured logging
				logx.Error("panic recovered", "err", fmt.Sprintf("%v", rec), "path", r.URL.Path, "method", r.Method)
				helperErr := fmt.Errorf("internal server error")
				web.Error(w, http.StatusInternalServerError, helperErr)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
