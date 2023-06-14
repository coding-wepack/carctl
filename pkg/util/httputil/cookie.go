package httputil

import "net/http"

func ParseCookie(cookie string) []*http.Cookie {
	header := http.Header{}
	header.Add("Cookie", cookie)
	req := http.Request{Header: header}
	return req.Cookies()
}

func GetXsrfToken(cookie string) string {
	cookies := ParseCookie(cookie)
	for _, c := range cookies {
		if c.Name == "XSRF-TOKEN" {
			return c.Value
		}
	}
	return ""
}
