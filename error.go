package rest

// These errors are handled when passing them into a method Context.Write and
// set the appropriate status response.
//
// Except this error, just checked that the error is responsible for
// os.IsNotExist (in this case, the status will be 404), or os.IsPermission
// (status 403). All other errors set the status to 500.
var (
	ErrBadRequest            = &Error{400, "bad request"}
	ErrUnauthorized          = &Error{401, "unauthorized"}
	ErrForbidden             = &Error{403, "forbidden"}
	ErrNotFound              = &Error{404, "not found"}
	ErrMethodNotAllowed      = &Error{405, "method not allowed"}
	ErrLengthRequired        = &Error{411, "length required"}
	ErrRequestEntityTooLarge = &Error{413, "request entity too large"}
	ErrUnsupportedMediaType  = &Error{415, "unsupported media type"}
	ErrInternalServerError   = &Error{500, "internal server error"}
	ErrMultipleResponse      = &Error{500, "multiple server response"}
	ErrNotImplemented        = &Error{501, "not implemented"}
	ErrServiceUnavailable    = &Error{503, "service unavailable"}
)

// Error describes the status and text message to send to a HTTP request.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"error"`
}

// Error returns a textual description of the error.
func (e *Error) Error() string {
	return e.Message
}
