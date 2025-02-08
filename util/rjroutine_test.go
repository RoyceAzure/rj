package util

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func makeData[T any](length int) []T {
	return make([]T, length)
}

/*
TODO 不要等到資料都處理完才寫入DB
*/
func TestSliceIterAndBroker(t *testing.T) {
	var num []int
	var wg sync.WaitGroup
	length := 5000
	var resultList []float32
	for i := 0; i < length; i++ {
		num = append(num, i)
	}
	batchSize := 15
	unporcessed := make(chan []int)
	processed := make(chan float32)
	wg.Add(4)
	go TaskDistributor(unporcessed, batchSize, num, nil, &wg)
	go TaskWorker("broker1", unporcessed, processed, processfun, nil, &wg)
	go TaskWorker("broker2", unporcessed, processed, processfun, nil, &wg)
	go TaskWorker("broker3", unporcessed, processed, processfun, nil, &wg)

	// go func() {
	// 	wg.Wait()
	// 	close(porcessed)
	// }()
	insertCount := 0
	go func() {
		wg.Wait()
		close(processed)
	}()
	for data := range processed {
		resultList = append(resultList, data)
		if len(resultList)%batchSize == 0 {
			insertCount += len(resultList)
			resultList = make([]float32, 0)
		}
	}
	if len(resultList) > 0 {
		insertCount += len(resultList)
	}
	require.Equal(t, length, insertCount)
}

func processfun(data int, parms ...any) (res float32, err error) {
	res = float32(data) + 0.01
	return res, nil
}

func TestTemp(t *testing.T) {
	parms := []interface{}{"spr_10:38:10"}

	// 反射来检查第一个元素的类型
	fmt.Println("Type of parms[0]:", reflect.TypeOf(parms[0]))

	// 尝试类型断言
	sprTime, ok := parms[0].(string)
	if !ok {
		fmt.Println("Type assertion to string failed.")
	} else {
		fmt.Println("Success:", sprTime)
	}
}
