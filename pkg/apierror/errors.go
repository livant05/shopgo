package apierror

import "net/http"

type Err struct {
	Code   string `json:"code"`
	Msg    string `json:"message"`
	Status int    `json:"-"`
	Cause  error  `json:"-"`
}

func (e *Err) Error() string { return e.Msg }
func (e *Err) Unwrap() error { return e.Cause }

func New(status int, code, msg string, cause error) *Err {
	return &Err{Code: code, Msg: msg, Status: status, Cause: cause}
}

var (
	BadRequest   = &Err{Code: "BAD_REQUEST", Msg: "solicitud inválida", Status: http.StatusBadRequest}
	Unauthorized = &Err{Code: "UNAUTHORIZED", Msg: "no autenticado", Status: http.StatusUnauthorized}
	Forbidden    = &Err{Code: "FORBIDDEN", Msg: "acceso denegado", Status: http.StatusForbidden}
	NotFound     = &Err{Code: "NOT_FOUND", Msg: "recurso no encontrado", Status: http.StatusNotFound}
	Conflict     = &Err{Code: "CONFLICT", Msg: "conflicto de datos", Status: http.StatusConflict}
	Internal     = &Err{Code: "INTERNAL", Msg: "error interno", Status: http.StatusInternalServerError}
)
