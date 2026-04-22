package httpruntime

import (
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc/status"
)

type errResponse struct {
	Error string `json:"error"`
}

// SetError is used to output errors to the client.
// You can override that in the runtime.
var SetError func(context.Context, *http.Request, http.ResponseWriter, error) = DefaultSetError

// DefaultSetError is the default error output.
func DefaultSetError(ctx context.Context, req *http.Request, w http.ResponseWriter, err error) {
	errCode := http.StatusInternalServerError
	if grpcErr, ok := err.(interface{ GRPCStatus() *status.Status }); ok {
		errCode = runtime.HTTPStatusFromCode(grpcErr.GRPCStatus().Code())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errCode)
	enc := json.NewEncoder(w)
	enc.Encode(errResponse{Error: err.Error()})
}
