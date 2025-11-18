package generic

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type RequestWithContext[C context.Context] http.Request

func (r *RequestWithContext[C]) Context() context.Context {
	ctx, ok := (*http.Request)(r).Context().(C)
	if !ok {
		var v C
		panic(fmt.Errorf("context type mismatch: expected %T, got %T", v, ctx))
	}
	return ctx
}

func (r *RequestWithContext[C]) AddCookie(c *http.Cookie) {
	(*http.Request)(r).AddCookie(c)
}

func (r *RequestWithContext[C]) BasicAuth() (username string, password string, ok bool) {
	return (*http.Request)(r).BasicAuth()
}

func (r *RequestWithContext[C]) Clone(ctx C) *RequestWithContext[C] {
	return (*RequestWithContext[C])((*http.Request)(r).Clone(ctx))
}

// Context method is already defined above

func (r *RequestWithContext[C]) Cookie(name string) (*http.Cookie, error) {
	return (*http.Request)(r).Cookie(name)
}

func (r *RequestWithContext[C]) Cookies() []*http.Cookie {
	return (*http.Request)(r).Cookies()
}

func (r *RequestWithContext[C]) CookiesNamed(name string) []*http.Cookie {
	return (*http.Request)(r).CookiesNamed(name)
}

func (r *RequestWithContext[C]) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return (*http.Request)(r).FormFile(key)
}

func (r *RequestWithContext[C]) FormValue(key string) string {
	return (*http.Request)(r).FormValue(key)
}

func (r *RequestWithContext[C]) MultipartReader() (*multipart.Reader, error) {
	return (*http.Request)(r).MultipartReader()
}

func (r *RequestWithContext[C]) ParseForm() error {
	return (*http.Request)(r).ParseForm()
}

func (r *RequestWithContext[C]) ParseMultipartForm(maxMemory int64) error {
	return (*http.Request)(r).ParseMultipartForm(maxMemory)
}

func (r *RequestWithContext[C]) PathValue(name string) string {
	return (*http.Request)(r).PathValue(name)
}

func (r *RequestWithContext[C]) PostFormValue(key string) string {
	return (*http.Request)(r).PostFormValue(key)
}

func (r *RequestWithContext[C]) ProtoAtLeast(major int, minor int) bool {
	return (*http.Request)(r).ProtoAtLeast(major, minor)
}

func (r *RequestWithContext[C]) Referer() string {
	return (*http.Request)(r).Referer()
}

func (r *RequestWithContext[C]) SetBasicAuth(username string, password string) {
	(*http.Request)(r).SetBasicAuth(username, password)
}

func (r *RequestWithContext[C]) SetPathValue(name string, value string) {
	(*http.Request)(r).SetPathValue(name, value)
}

func (r *RequestWithContext[C]) UserAgent() string {
	return (*http.Request)(r).UserAgent()
}

func (r *RequestWithContext[C]) WithContext(ctx context.Context) *RequestWithContext[C] {
	return (*RequestWithContext[C])((*http.Request)(r).WithContext(ctx))
}

func (r *RequestWithContext[C]) Write(w io.Writer) error {
	return (*http.Request)(r).Write(w)
}

func (r *RequestWithContext[C]) WriteProxy(w io.Writer) error {
	return (*http.Request)(r).WriteProxy(w)
}

func NewRequestWithContext[C context.Context](ctx C, method string, url string, body io.Reader) (*RequestWithContext[C], error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	return (*RequestWithContext[C])(req), nil
}
