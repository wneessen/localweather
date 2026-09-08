// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package server

import (
	"net/http"
	"time"

	"github.com/go-chi/render"

	"github.com/wneessen/localweather/internal/apperror"
	"github.com/wneessen/localweather/internal/log"
)

type Response struct {
	Success    bool          `json:"success"`
	StatusCode int           `json:"statusCode"`
	Status     string        `json:"status"`
	Message    string        `json:"message,omitempty"`
	Timestamp  time.Time     `json:"timestamp"`
	RequestID  string        `json:"requestId,omitempty"`
	Data       any           `json:"data,omitempty"`
	Errors     []ErrorDetail `json:"errors,omitempty"`
}

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Render satisfies the go-chi render.Renderer interface.
func (re *Response) Render(_ http.ResponseWriter, r *http.Request) error {
	if re.StatusCode != 0 {
		render.Status(r, re.StatusCode)
	}
	if re.Timestamp.IsZero() {
		re.Timestamp = time.Now().UTC()
	}
	return nil
}

func NewResponse(code int, msg string, data any) *Response {
	return &Response{
		Success:    true,
		StatusCode: code,
		Status:     http.StatusText(code),
		Message:    msg,
		Timestamp:  time.Now().UTC(),
		Data:       data,
	}
}

// NewError builds an error envelope with optional details.
func NewError(code int, msg string, details ...ErrorDetail) *Response {
	return &Response{
		Success:    false,
		StatusCode: code,
		Status:     http.StatusText(code),
		Message:    msg,
		Timestamp:  time.Now().UTC(),
		Errors:     details,
	}
}

func (s *Server) renderErr(w http.ResponseWriter, r *http.Request, status int, err error) {
	if rerr := render.Render(w, r, NewError(status, err.Error())); rerr != nil {
		s.log.Error(apperror.ErrFailedToRenderErrorResponse.Error(), log.ErrAttr(rerr))
	}
}
