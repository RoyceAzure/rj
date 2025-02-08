package error_ana

import "fmt"

type ErrCode int

const (
	// 5xx Server Errors
	InternalErrorCode   ErrCode = 500 // 內部服務錯誤
	UnavailableCode     ErrCode = 503 // 服務不可用
	ServiceTimeoutCode  ErrCode = 504 // 服務超時
	DatabaseErrorCode   ErrCode = 520 // 數據庫錯誤
	CacheErrorCode      ErrCode = 521 // 緩存錯誤
	ThirdPartyErrorCode ErrCode = 522 // 第三方服務錯誤

	// 4xx Client Errors
	BadRequestCode       ErrCode = 400 // 請求格式錯誤
	UnauthenticatedCode  ErrCode = 401 // 未認證
	UnauthorizedCode     ErrCode = 403 // 未授權
	NotFoundCode         ErrCode = 404 // 資源未找到
	MethodNotAllowedCode ErrCode = 405 // 請求方法不允許
	ConflictCode         ErrCode = 409 // 資源衝突
	RequestTimeoutCode   ErrCode = 408 // 請求超時
	TooManyRequestsCode  ErrCode = 429 // 請求過於頻繁

	// 業務邏輯錯誤 (可以用 4xx 範圍)
	InvalidArgumentCode   ErrCode = 460 // 參數錯誤
	InvalidFormatCode     ErrCode = 461 // 格式錯誤
	DataNotExistsCode     ErrCode = 462 // 數據不存在
	DataAlreadyExistsCode ErrCode = 463 // 數據已存在
	InvalidOperationCode  ErrCode = 464 // 非法操作
	BusinessErrorCode     ErrCode = 465 // 業務邏輯錯誤

	// 用戶相關錯誤
	UserNotFoundCode    ErrCode = 470 // 用戶不存在
	UserDisabledCode    ErrCode = 471 // 用戶已禁用
	UserLockedCode      ErrCode = 472 // 用戶已鎖定
	InvalidPasswordCode ErrCode = 473 // 密碼錯誤
	TokenExpiredCode    ErrCode = 474 // Token過期
	InvalidTokenCode    ErrCode = 475 // Token無效
)

var ErrStrMap = map[ErrCode]string{
	// 5xx
	InternalErrorCode:   "internal server error",
	UnavailableCode:     "service unavailable",
	ServiceTimeoutCode:  "service timeout",
	DatabaseErrorCode:   "database error",
	CacheErrorCode:      "cache error",
	ThirdPartyErrorCode: "third party service error",

	// 4xx
	BadRequestCode:       "bad request",
	UnauthenticatedCode:  "unauthenticated",
	UnauthorizedCode:     "unauthorized",
	NotFoundCode:         "data not found",
	MethodNotAllowedCode: "method not allowed",
	ConflictCode:         "resource conflict",
	RequestTimeoutCode:   "request timeout",
	TooManyRequestsCode:  "too many requests",

	// 業務邏輯
	InvalidArgumentCode:   "invalid argument",
	InvalidFormatCode:     "invalid format",
	DataNotExistsCode:     "data not exists",
	DataAlreadyExistsCode: "data already exists",
	InvalidOperationCode:  "invalid operation",
	BusinessErrorCode:     "business logic error",

	// 用戶相關
	UserNotFoundCode:    "user not found",
	UserDisabledCode:    "user is disabled",
	UserLockedCode:      "user is locked",
	InvalidPasswordCode: "invalid password",
	TokenExpiredCode:    "token expired",
	InvalidTokenCode:    "invalid token",
}

var (
	// 5xx Server Errors
	InternalError       = &AnaError{Code: InternalErrorCode, CodeMsg: ErrStrMap[InternalErrorCode]}
	UnavailableError    = &AnaError{Code: UnavailableCode, CodeMsg: ErrStrMap[UnavailableCode]}
	ServiceTimeoutError = &AnaError{Code: ServiceTimeoutCode, CodeMsg: ErrStrMap[ServiceTimeoutCode]}
	DatabaseError       = &AnaError{Code: DatabaseErrorCode, CodeMsg: ErrStrMap[DatabaseErrorCode]}
	CacheError          = &AnaError{Code: CacheErrorCode, CodeMsg: ErrStrMap[CacheErrorCode]}
	ThirdPartyError     = &AnaError{Code: ThirdPartyErrorCode, CodeMsg: ErrStrMap[ThirdPartyErrorCode]}

	// 4xx Client Errors
	BadRequestError       = &AnaError{Code: BadRequestCode, CodeMsg: ErrStrMap[BadRequestCode]}
	UnauthenticatedError  = &AnaError{Code: UnauthenticatedCode, CodeMsg: ErrStrMap[UnauthenticatedCode]}
	UnauthorizedError     = &AnaError{Code: UnauthorizedCode, CodeMsg: ErrStrMap[UnauthorizedCode]}
	NotFoundError         = &AnaError{Code: NotFoundCode, CodeMsg: ErrStrMap[NotFoundCode]}
	MethodNotAllowedError = &AnaError{Code: MethodNotAllowedCode, CodeMsg: ErrStrMap[MethodNotAllowedCode]}
	ConflictError         = &AnaError{Code: ConflictCode, CodeMsg: ErrStrMap[ConflictCode]}
	RequestTimeoutError   = &AnaError{Code: RequestTimeoutCode, CodeMsg: ErrStrMap[RequestTimeoutCode]}
	TooManyRequestsError  = &AnaError{Code: TooManyRequestsCode, CodeMsg: ErrStrMap[TooManyRequestsCode]}

	// Business Logic Errors
	InvalidArgumentError  = &AnaError{Code: InvalidArgumentCode, CodeMsg: ErrStrMap[InvalidArgumentCode]}
	InvalidFormatError    = &AnaError{Code: InvalidFormatCode, CodeMsg: ErrStrMap[InvalidFormatCode]}
	DataNotExistsError    = &AnaError{Code: DataNotExistsCode, CodeMsg: ErrStrMap[DataNotExistsCode]}
	DataExistsError       = &AnaError{Code: DataAlreadyExistsCode, CodeMsg: ErrStrMap[DataAlreadyExistsCode]}
	InvalidOperationError = &AnaError{Code: InvalidOperationCode, CodeMsg: ErrStrMap[InvalidOperationCode]}
	BusinessError         = &AnaError{Code: BusinessErrorCode, CodeMsg: ErrStrMap[BusinessErrorCode]}

	// User Related Errors
	UserNotFoundError    = &AnaError{Code: UserNotFoundCode, CodeMsg: ErrStrMap[UserNotFoundCode]}
	UserDisabledError    = &AnaError{Code: UserDisabledCode, CodeMsg: ErrStrMap[UserDisabledCode]}
	UserLockedError      = &AnaError{Code: UserLockedCode, CodeMsg: ErrStrMap[UserLockedCode]}
	InvalidPasswordError = &AnaError{Code: InvalidPasswordCode, CodeMsg: ErrStrMap[InvalidPasswordCode]}
	TokenExpiredError    = &AnaError{Code: TokenExpiredCode, CodeMsg: ErrStrMap[TokenExpiredCode]}
	InvalidTokenError    = &AnaError{Code: InvalidTokenCode, CodeMsg: ErrStrMap[InvalidTokenCode]}
)

type AnaError struct {
	Code        ErrCode
	CodeMsg     string
	ExternalMsg string
}

func New(code ErrCode, externalMsg string) *AnaError {
	return &AnaError{
		Code:        code,
		CodeMsg:     ErrStrMap[code],
		ExternalMsg: externalMsg,
	}
}

func (err *AnaError) Error() string {
	if len(err.ExternalMsg) == 0 {
		return fmt.Sprint(ErrStrMap[err.Code])
	}
	return fmt.Sprintf("%s, %s", ErrStrMap[err.Code], err.ExternalMsg)
}

func (err *AnaError) Is(target error) bool {
	ana_err, ok := target.(*AnaError)
	if !ok {
		return false
	}
	return err.Code == ana_err.Code
}
