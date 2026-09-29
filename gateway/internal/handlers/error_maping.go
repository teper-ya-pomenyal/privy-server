package handlers

import (
	"net/http"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ValidationError struct {
	Code    string `json:"code"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

func writeValidationError(w http.ResponseWriter, field, code, message string) {
	writeJSON(w, http.StatusBadRequest, ValidationError{Code: code, Field: field, Message: message})
}

func mapGRPCError(w http.ResponseWriter, err error) {
	st := status.Convert(err)

	// field-level validation error from the service
	if st.Code() == codes.InvalidArgument {
		for _, d := range st.Details() {
			if br, ok := d.(*errdetails.BadRequest); ok && len(br.FieldViolations) > 0 {
				v := br.FieldViolations[0]
				writeValidationError(w, v.Field, v.Reason, st.Message())
				return
			}
		}
	}

	var httpStatus int

	switch st.Code() {
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
	case codes.AlreadyExists:
		httpStatus = http.StatusConflict
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	default:
		httpStatus = http.StatusInternalServerError
	}

	http.Error(w, st.Message(), httpStatus)
}
