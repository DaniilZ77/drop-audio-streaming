package grpc

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func withDetails(code codes.Code, err error, details map[string]string) error {
	st := status.New(code, err.Error())
	var violations []*errdetails.QuotaFailure_Violation
	for k, v := range details {
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
