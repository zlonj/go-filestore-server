package config

const (
	AsyncTransferEnable  = false
	RabbitURL            = "amqp://guest:guest@127.0.0.1:5672/"
	TransExchangeName    = "filestore-server-s3"
	TransS3QueueName    = "filestore-server-s3-queue"
	TransS3SErrQueueName = "filestore-server-s3-queue-err"
	TransS3RoutingKey   = "s3"
)
