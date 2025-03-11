package rj_error

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ConvertGrpcErrorToAppError 将 gRPC 错误转换为应用错误
func ConvertGrpcErrorToAppError(err error) *AnaError {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return New(InternalErrorCode, err.Error())
	}

	message := st.Message()

	switch st.Code() {
	// 4xx - 客户端错误
	case codes.InvalidArgument:
		return New(InvalidArgumentCode, message)
	case codes.NotFound:
		return New(NotFoundCode, message)
	case codes.AlreadyExists:
		return New(DataAlreadyExistsCode, message)
	case codes.PermissionDenied:
		return New(UnauthorizedCode, message)
	case codes.Unauthenticated:
		return New(UnauthenticatedCode, message)
	case codes.FailedPrecondition:
		return New(InvalidOperationCode, message)
	case codes.OutOfRange:
		return New(InvalidArgumentCode, message)
	case codes.Unimplemented:
		return New(MethodNotAllowedCode, message)
	case codes.Canceled:
		return New(RequestTimeoutCode, message)
	case codes.DeadlineExceeded:
		return New(ServiceTimeoutCode, message)
	case codes.ResourceExhausted:
		return New(TooManyRequestsCode, message)
	case codes.Aborted:
		return New(ConflictCode, message)

	// 5xx - 服务器错误
	case codes.Internal:
		return New(InternalErrorCode, message)
	case codes.Unavailable:
		return New(UnavailableCode, message)
	case codes.DataLoss:
		return New(DatabaseErrorCode, message)
	case codes.Unknown:
		return New(InternalErrorCode, message)

	// 默认处理
	default:
		return New(InternalErrorCode, message)
	}
}
