package test

import (
	"errors"

	"testing"

	er "github.com/RoyceAzure/rj/util/rj_error"
)

func TestCustomError(t *testing.T) {
	// 創建兩個相同代碼但不同消息的錯誤
	err1 := er.New(er.BadRequestCode, "參數格式錯誤")

	err2 := er.New(er.BadRequestCode, "缺少參數")

	// 測試 errors.Is 比較
	if !errors.Is(err1, err2) {
		t.Error("相同Code的錯誤應該被視為相同類型")
	}

	// 創建不同代碼的錯誤
	err3 := er.New(er.NotFoundCode, "Resource not found")

	// 測試不同代碼的錯誤比較
	if errors.Is(err1, err3) {
		t.Error("不同Code的錯誤不應該被視為相同類型")
	}

	//測試比較面板值
	if !errors.Is(err1, er.BadRequestError) {
		t.Error("相同Code的錯誤應該被視為相同類型")
	}

	// 測試錯誤包裝
	wrappedErr := er.WrapError(err1, "API call failed")
	if !errors.Is(wrappedErr, err1) {
		t.Error("包裝後的錯誤應該能夠識別原始錯誤")
	}
}

// 測試錯誤代碼比較
func TestErrorCodeComparison(t *testing.T) {
	// 創建一個基準錯誤
	baseErr := er.New(er.ServiceTimeoutCode, "Operation timed out")

	// 使用不同方式創建錯誤進行比較
	testCases := []struct {
		name      string
		testErr   error
		plainErr  error
		targetErr error
		expected  bool
	}{
		{
			name:      "Same error code",
			testErr:   er.New(er.ServiceTimeoutCode, "不同超時訊息"),
			plainErr:  er.ServiceTimeoutError,
			targetErr: baseErr,
			expected:  true,
		},
		{
			name:      "Different error code",
			testErr:   er.New(er.NotFoundCode, "找不到資源"),
			plainErr:  er.NotFoundError,
			targetErr: baseErr,
			expected:  false,
		},
		{
			name:      "Wrapped error",
			testErr:   er.WrapError(baseErr, "包裝超時訊息"),
			plainErr:  er.ServiceTimeoutError,
			targetErr: baseErr,
			expected:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := errors.Is(tc.testErr, tc.targetErr)
			plainResult := errors.Is(tc.plainErr, tc.targetErr)
			if result != tc.expected {
				t.Errorf("Test case %s failed: expected %v but got %v",
					tc.name, tc.expected, result)
			}
			if plainResult != tc.expected {
				t.Errorf("Test case %s failed: expected %v but got %v",
					tc.name, tc.expected, plainResult)
			}
		})
	}
}
