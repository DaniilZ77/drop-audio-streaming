package grpc

import (
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model/validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(v *validator.Validator) error {
	st := status.New(codes.InvalidArgument, core.ErrValidationFailed.Error())
	var violations []*errdetails.QuotaFailure_Violation
	for k, v := range v.Errors {
		violations = append(violations, &errdetails.QuotaFailure_Violation{
			Subject:     k,
			Description: v,
		})
	}
	ds, err := st.WithDetails(&errdetails.QuotaFailure{Violations: violations})
	if err != nil {
		return st.Err()
	}
	return ds.Err()
}
