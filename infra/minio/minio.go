package minio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/xuri/excelize/v2"
)

// 線程安全  可共用
type MinioClient struct {
	client *minio.Client
}

func NewMinioClient(endpoint, port, accessKey, secretKey string) (*MinioClient, error) {
	// 創建 MinIO client
	client, err := minio.New(fmt.Sprintf("%s:%s", endpoint, port), &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false, // 本地開發用 false
	})
	if err != nil {
		return nil, err
	}

	return &MinioClient{
		client: client,
	}, nil
}

func (m *MinioClient) GetObject(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	if bucketName == "" || objectName == "" {
		return nil, fmt.Errorf("bucket name and object name cannot be empty")
	}

	exists, err := m.client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("bucket '%s' does not exist", bucketName)
	}

	// Get object info to verify existence and get size
	info, err := m.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	// Get the object
	object, err := m.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	// Read the object data
	data := make([]byte, info.Size)
	_, err = io.ReadFull(object, data)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}

	return data, nil
}

// ListObjects 列出指定 bucket 的所有物件
func (m *MinioClient) ListObjects(ctx context.Context, bucketName string, isRecursive bool) ([]string, error) {
	var objects []string

	objectCh := m.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Recursive: isRecursive,
	})

	for object := range objectCh {
		if object.Err != nil {
			//todo log warning
			continue
		}
		objects = append(objects, object.Key)
	}

	return objects, nil
}

// ListObjectsWithPrefix 用前綴列出物件
func (m *MinioClient) ListObjectsWithPrefix(ctx context.Context, bucketName, prefix string) ([]string, error) {
	var objects []string

	objectCh := m.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			//todo log warning
			continue
		}
		objects = append(objects, object.Key)
	}

	return objects, nil
}

func (m *MinioClient) DownloadFile(ctx context.Context, bucketName, objectName string, fileType string) ([]byte, error) {
	// 獲取物件
	object, err := m.client.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object error: %w", err)
	}
	defer object.Close()

	// 讀取內容
	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("read object error: %w", err)
	}

	// 依據檔案類型處理
	switch fileType {
	case "json":
		// 確認是否為有效的 JSON
		if !json.Valid(data) {
			return nil, fmt.Errorf("invalid JSON file")
		}
	case "xlsx":
		// 確認是否為有效的 Excel 文件
		_, err := excelize.OpenReader(object)
		if err != nil {
			return nil, fmt.Errorf("invalid Excel file: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}

	return data, nil
}

// func (m *MinioClient) DownloadExcel(bucketName, objectName string) (*excelize.File, error) {
// 	// 使用系統臨時目錄
// 	tempDir := os.TempDir()
// 	tempFileName := filepath.Join(tempDir, fmt.Sprintf("temp-%d.xlsx", time.Now().UnixNano()))

// 	// 下載檔案
// 	err := m.client.FGetObject(
// 		context.Background(),
// 		bucketName,
// 		objectName,
// 		tempFileName,
// 		minio.GetObjectOptions{},
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("download file error: %v", err)
// 	}

// 	// 確保清理臨時檔案
// 	defer os.Remove(tempFileName)

// 	// 開啟 Excel
// 	data, err := io.ReadAll(tempFileName)
// 	if err != nil {
// 		return nil, fmt.Errorf("read object error: %v", err)
// 	}
// 	f, err := excelize.OpenReader(tempFileName)
// 	if err != nil {
// 		return nil, fmt.Errorf("open excel error: %v", err)
// 	}

// 	return f, nil
// }
