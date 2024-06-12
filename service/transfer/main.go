package transfer

import (
	"context"
	"encoding/json"
	"filestore-server/config"
	"filestore-server/meta"
	"filestore-server/mq"
	appS3 "filestore-server/store/s3"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func ProcessTransfer(msg []byte) bool {
	pubData := mq.TransferData{}
	err := json.Unmarshal(msg, &pubData)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	file, err := os.Open(pubData.CurrLocation)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	// Store file to S3
	s3Bucket := config.S3_BUCKET
	s3Key := pubData.FileHash
	_, err = appS3.Client().PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &s3Bucket,
		Key: &s3Key,
		Body: file,
	})
	if err != nil {
		fmt.Println("Upload S3 err: " + err.Error())
		return false
	}

	// Update file metadata location
	fileMeta := meta.GetFileMeta(pubData.FileHash)
	fileMeta.Location = pubData.DestLocation
	suc := meta.UpdateFileMetadataDB(fileMeta)
	return suc
}

func main() {
	fmt.Println("Listening for message queue")
	mq.StartConsume(config.TransS3QueueName, "transfer_s3", ProcessTransfer)
}