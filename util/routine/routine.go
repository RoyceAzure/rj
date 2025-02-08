package routine

import "sync"

/*
將資料已batchSize的大小分配slice，並儲存到ch裡面
分配資料完畢時關閉ch
由於有關閉通道行為，所以只能由一個goroutine啟動
*/
func TaskDistributor[T any](ch chan<- []T, batchSize int, target []T, fiterFuncList []func([]T) []T, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(ch)
	if length := len(fiterFuncList); length != 0 {
		for i := 0; i < length; i++ {
			target = fiterFuncList[i](target)
		}
	}
	for i := 0; i < len(target); i += batchSize {
		end := i + batchSize
		if end > len(target) {
			end = len(target)
		}
		ch <- target[i:end]
	}
}

/*
從unporcessed chan 接收[]資料
由processFunc 處理資料
最後儲存到porcessed chan

defer wg.Done()
*/
func TaskWorker[T any, T1 any](name string, unprocessed <-chan []T,
	porcessed chan<- T1,
	processFunc func(data T, parms ...any) (T1, error),
	errorFunc func(error),
	wg *sync.WaitGroup,
	parms ...any) {
	defer wg.Done()
	for dataBatch := range unprocessed {
		for _, data := range dataBatch {
			res, err := processFunc(data, parms...)
			if err != nil {
				if errorFunc != nil {
					errorFunc(err)
				}
				continue
			}
			porcessed <- res
		}
	}
}
