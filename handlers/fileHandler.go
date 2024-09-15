package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"
    "log"

    "bytes"
	"context"
	"io"
	"mime/multipart"
	"sync"
    "gorm.io/gorm"
    "mime"

    "21BCE2661_Backend/db"
    "21BCE2661_Backend/models"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/gin-gonic/gin"
)

type FileMetadata struct {
    ID         int    `json:"id"`
    FileName   string `json:"file_name"`
    UploadDate time.Time `json:"upload_date"`
    FileSize   int64  `json:"file_size"`
    FileURL    string `json:"file_url"`
    FileType   string `json:"file_type"`
}

func UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
		return
	}
	defer file.Close()

    region :=os.Getenv("AWS_REGION")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create AWS session"})
		return
	}

	svc := s3.New(sess)

	bucket := os.Getenv("S3_BUCKET")
	objectKey := fmt.Sprintf("uploads/%s", header.Filename)

	upload, err := svc.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create multipart upload"})
		return
	}

	const chunkSize = 5 * 1024 * 1024 

	type UploadPart struct {
		PartNumber int64
		ETag       *string
		Err        error
	}

	ch := make(chan UploadPart)
	var wg sync.WaitGroup
	partNumber := int64(1)

	for {
		buffer := make([]byte, chunkSize)
		bytesRead, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading file"})
			return
		}
		if bytesRead == 0 {
			break
		}

		chunk := buffer[:bytesRead]

		wg.Add(1)
		go func(partNumber int64, chunk []byte) {
			defer wg.Done()
			partOutput, err := svc.UploadPart(&s3.UploadPartInput{
				Bucket:     aws.String(bucket),
				Key:        aws.String(objectKey),
				UploadId:   upload.UploadId,
				PartNumber: aws.Int64(partNumber),
				Body:       bytes.NewReader(chunk),
			})
			ch <- UploadPart{
				PartNumber: partNumber,
				ETag:       partOutput.ETag,
				Err:        err,
			}
		}(partNumber, chunk)

		partNumber++
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var completedParts []*s3.CompletedPart
	for part := range ch {
		if part.Err != nil {
			svc.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
				Bucket:   aws.String(bucket),
				Key:      aws.String(objectKey),
				UploadId: upload.UploadId,
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload part %d", part.PartNumber)})
			return
		}
		completedParts = append(completedParts, &s3.CompletedPart{
			ETag:       part.ETag,
			PartNumber: aws.Int64(part.PartNumber),
		})
	}

	_, err = svc.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(objectKey),
		UploadId: upload.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete multipart upload"})
		return
	}

	fileURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, objectKey)

    fileType := getFileType(file)

	email, exists := c.Get("email")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }


    fileMetadata := models.File{
		Email:    	email.(string),
        FileName:   header.Filename,
        FileURL:    fileURL,
        UploadDate: time.Now(),
        FileSize:   header.Size,
        FileType:   fileType,
    }

    result := db.DB.Create(&fileMetadata)
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file metadata"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "File uploaded successfully",
        "file_url": fileURL,
    })
}

func GetFiles(c *gin.Context) {
    email, exists := c.Get("email")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    cacheKey := fmt.Sprintf("files:%s", email)
    cachedFiles, err := db.RedisClient.Get(c.Request.Context(), cacheKey).Result()
    if err == nil && cachedFiles != "" {
        var fileResponses []FileMetadata
        if err := json.Unmarshal([]byte(cachedFiles), &fileResponses); err == nil {
            c.JSON(http.StatusOK, gin.H{"files": fileResponses})
            return
        }
    }

    var files []models.File
    result := db.DB.Where("email = ?", email).Find(&files)
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
        return
    }

    var fileResponses []FileMetadata
    for _, file := range files {
        fileResponses = append(fileResponses, FileMetadata{
            ID:         file.ID,
            FileName:   file.FileName,
            UploadDate: file.UploadDate,
            FileSize:   file.FileSize,
            FileURL:    file.FileURL,
            FileType:   file.FileType,
        })
    }

    fileResponsesJSON, err := json.Marshal(fileResponses)
    if err == nil {
        db.RedisClient.Set(c.Request.Context(), cacheKey, fileResponsesJSON, time.Minute*5)
    }

    c.JSON(http.StatusOK, gin.H{"files": fileResponses})
}

func GetFileURL(c *gin.Context) {
    fileID := c.Param("file_id")

    cacheKey := fmt.Sprintf("file:%s", fileID)
    cachedFileURL, err := db.RedisClient.Get(c.Request.Context(), cacheKey).Result()
    if err == nil && cachedFileURL != "" {
        c.JSON(http.StatusOK, gin.H{"file_url": cachedFileURL})
        return
    }

    var file models.File
    result := db.DB.Take(&file, "id = ?", fileID)
    if result.Error != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
        return
    }

    db.RedisClient.Set(c.Request.Context(), cacheKey, file.FileURL, time.Minute*5)

    c.JSON(http.StatusOK, gin.H{"file_url": file.FileURL})
}


func SearchFiles(c *gin.Context) {
    name := c.Query("name")
    uploadDate := c.Query("upload_date")
    fileType := c.Query("file_type")

    var query *gorm.DB
    var cacheKey string
    var files []models.File

    switch {
    case name != "":
        cacheKey = fmt.Sprintf("search:name:%s", name)
        query = db.DB.Model(&models.File{}).Where("file_name ILIKE ?", "%"+name+"%")
    case uploadDate != "":
        cacheKey = fmt.Sprintf("search:date:%s", uploadDate)
        query = db.DB.Model(&models.File{}).Where("upload_date = ?", uploadDate)
    case fileType != "":
        cacheKey = fmt.Sprintf("search:type:%s", fileType)
        query = db.DB.Model(&models.File{}).Where("file_type = ?", fileType)
    default:
        c.JSON(http.StatusBadRequest, gin.H{"error": "No search parameter provided"})
        return
    }

    cachedFiles, err := db.RedisClient.Get(c.Request.Context(), cacheKey).Result()
    if err == nil && cachedFiles != "" {
        if err := json.Unmarshal([]byte(cachedFiles), &files); err == nil {
            c.JSON(http.StatusOK, gin.H{"files": files})
            return
        }
    }

    result := query.Find(&files)
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
        return
    }

    filesJSON, err := json.Marshal(files)
    if err == nil {
        db.RedisClient.Set(c.Request.Context(), cacheKey, filesJSON, 5*time.Minute)
    }

    c.JSON(http.StatusOK, gin.H{"files": files})
}


func getFileType(file multipart.File) string {
    buffer := make([]byte, 512) 
    _, err := file.Read(buffer)
    if err != nil {
        return "unknown"
    }

    return mime.TypeByExtension(http.DetectContentType(buffer))
}

func DeleteExpiredFiles() {
    var files []models.File
    now := time.Now()
    
    db.DB.Where("upload_date < ?", now.Add(-30*24*time.Hour)).Find(&files)

    region := os.Getenv("AWS_REGION")
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(region),
    })
    if err != nil {
        log.Fatalf("failed to create session: %v", err)
    }
    
    s3Client := s3.New(sess)

    bucket := os.Getenv("S3_BUCKET")

    for _, file := range files {
        _, err := s3Client.DeleteObject(&s3.DeleteObjectInput{
            Bucket: aws.String(bucket), 
            Key:    aws.String(file.FileURL),
        })
        if err != nil {
            fmt.Printf("Failed to delete file from S3: %v\n", err)
            continue
        }

        db.DB.Delete(&file)
    }

    ctx := context.Background()

    errRed := db.RedisClient.FlushDB(ctx).Err() 
    if errRed != nil {
        fmt.Printf("Failed to clear Redis cache: %v\n", errRed)
    }
}

func StartWorker() {
    ticker := time.NewTicker(24 * time.Hour) 
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            DeleteExpiredFiles()
        }
    }
}