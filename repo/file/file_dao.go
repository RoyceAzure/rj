package file

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

type FileDAO interface {
	Write(content string) error
	Append(content string) error
	Read() ([]string, error)
	Clear() error
	Close() error
}

type TxtFileDAO struct {
	filePath     string
	mu           sync.RWMutex
	appendFile   *os.File      //for logger append
	appendWriter *bufio.Writer //for logger append
}

// NewTxtFileDAO 創建新的文件操作實例
func NewTxtFileDAO(filePath string) (*TxtFileDAO, error) {
	// 檢查文件是否存在，不存在則創建
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	return &TxtFileDAO{
		filePath:     filePath,
		appendFile:   file,
		appendWriter: bufio.NewWriter(file),
	}, nil
}

func (dao *TxtFileDAO) Write(content string) error {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	file, err := os.OpenFile(dao.filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for writing: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write content: %v", err)
	}

	// 確保所有數據都被寫入
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush writer: %v", err)
	}

	return nil
}

// Append 追加內容到文件末尾
func (dao *TxtFileDAO) Append(content string) error {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	file, err := os.OpenFile(dao.filePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for appending: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(content + "\n")
	if err != nil {
		return fmt.Errorf("failed to append content: %v", err)
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush writer: %v", err)
	}

	return nil
}

// Read 讀取文件全部內容
func (dao *TxtFileDAO) Read() ([]string, error) {
	// 先取得寫入鎖，確保寫入緩衝區被刷新
	dao.mu.Lock()
	// 刷新緩衝區，確保所有數據都寫入檔案
	if err := dao.appendWriter.Flush(); err != nil {
		dao.mu.Unlock()
		return nil, fmt.Errorf("failed to flush writer: %v", err)
	}
	dao.mu.Unlock()

	dao.mu.RLock()
	defer dao.mu.RUnlock()

	file, err := os.OpenFile(dao.filePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for reading: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error while reading file: %v", err)
	}

	return lines, nil
}

// Clear 清空文件內容
func (dao *TxtFileDAO) Clear() error {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	file, err := os.OpenFile(dao.filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to clear file: %v", err)
	}
	defer file.Close()

	return nil
}

// Close 關閉文件
func (dao *TxtFileDAO) Close() error {
	dao.mu.Lock()
	defer dao.mu.Unlock()

	if dao.appendWriter != nil {
		if err := dao.appendWriter.Flush(); err != nil {
			return fmt.Errorf("failed to flush writer: %v", err)
		}
	}

	if dao.appendFile != nil {
		if err := dao.appendFile.Close(); err != nil {
			return fmt.Errorf("failed to close file: %v", err)
		}
	}

	return nil
}
