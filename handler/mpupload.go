package handler

import (
	"filestore-server/util"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	redisPool "filestore-server/cache/redis"
)

type MultipartUploadInfo struct {
	FileHash string
	FileSize int
	UploadID string
	ChunkSize int
	ChunkCount int
}

// Initialize multi-part upload
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request params
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}
	// 2. Get redis pool connection
	redisConn := redisPool.RedisPool().Get()
	defer redisConn.Close()

	// 3. Initialize multipart upload information
	upInfo := MultipartUploadInfo{
		FileHash: filehash,
		FileSize: filesize,
		UploadID: username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize: 5 * 1024 * 1024,
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}

	// 4. Write multipart information into redis
	redisConn.Do("HSET", "MP_" + upInfo.UploadID, "chunkcount", upInfo.ChunkCount)
	redisConn.Do("HSET", "MP_" + upInfo.UploadID, "filehash", upInfo.FileHash)
	redisConn.Do("HSET", "MP_" + upInfo.UploadID, "filesize", upInfo.FileSize)

	// 5. Return response back to client
	w.Write(util.NewRespMsg(0, "OK", upInfo).JSONBytes())
}