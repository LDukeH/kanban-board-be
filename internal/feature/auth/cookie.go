package auth

import (
	"net/http"
	"net/url"
	"os"
	"strings"
)

func sameSiteMode() http.SameSite {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("COOKIE_SAMESITE"))) {
	case "none":
		return http.SameSiteNoneMode
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	}

	frontend := os.Getenv("FRONTEND_URL")
	if frontend == "" {
		return http.SameSiteLaxMode
	}
	if u, err := url.Parse(frontend); err == nil && u.Hostname() != "" && !isLocalHost(u.Hostname()) {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func isLocalHost(host string) bool {
	host = strings.ToLower(host)
	return host == "localhost" || host == "127.0.0.1" || host == "::1" || strings.HasSuffix(host, ".localhost")
}

func secureCookie() bool {
	return sameSiteMode() == http.SameSiteNoneMode
}

func newAccessTokenCookie(token string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secureCookie(),
		SameSite: sameSiteMode(),
	}
}

func clearAccessTokenCookie() *http.Cookie {
	return newAccessTokenCookie("", -1)
}
