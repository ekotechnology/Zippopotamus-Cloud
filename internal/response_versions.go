package internal

import (
	"context"
	"net/http"
	"strings"
)

func SetVersion(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		requestedVersion := r.Header.Get("ZP-Version")

		requestedVersion = strings.TrimSpace(requestedVersion)

		if requestedVersion == "" {
			r.ParseForm()

			v := r.Form.Get("zp-version")

			if v == "" {
				requestedVersion = "v1"
			} else {
				v = strings.TrimSpace(v)

				if v == "" {
					requestedVersion = "v1"
				} else {
					requestedVersion = v
				}
			}
		}

		r = r.WithContext(context.WithValue(ctx, "ZPVersion", requestedVersion))

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func GetVersion(r *http.Request) string {
	version, ok := r.Context().Value("ZPVersion").(string)

	if !ok {
		return ""
	}

	switch version {
	case "v1":
		return version
	case "v2":
		return version
	default:
		return "v1"
	}
}
