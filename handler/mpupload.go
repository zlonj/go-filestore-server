package handler

import (
	dblayer "filestore-server/db"
	"filestore-server/util"
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	redisPool "filestore-server/cache/redis"

	"github.com/gomodule/redigo/redis"
)

type MultipartUploadInfo struct {
	FileHash string
	FileSize int
	UploadID string
	ChunkSize int
	ChunkCount int
	ChunkExists []int
}

const (
	ChunkDir = "/data/chunks/"
	MergeDir = "/data/merge/"
	ChunkKeyPrefix = "MP_"
	HashUpIDKeyPrefix = "HASH_UPID_"
)

func init() {
	if err := os.MkdirAll(ChunkDir, 0744); err != nil {
		fmt.Println("Cannot mkdir: " + ChunkDir)
		os.Exit(1)
	}

	if err := os.MkdirAll(MergeDir, 0744); err != nil {
		fmt.Println("Cannot mkdir: " + MergeDir)
		os.Exit(1)
	}
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

	// 3. Check if any parts are already uploaded
	uploadID := ""
	keyExists, _ := redis.Bool(redisConn.Do("EXISTS", HashUpIDKeyPrefix + filehash))
	if keyExists {
		uploadID, err = redis.String(redisConn.Do("GET", HashUpIDKeyPrefix + filehash))
		if err != nil {
			w.Write(util.NewRespMsg(-1, "Upload part failed", err.Error()).JSONBytes())
			return
		}
	}

	// 4.1 First upload: create new uploadID
	// 4.2 If continue to upload, get chunks already uploaded with uploadID
	chunksExist := []int{}

	if uploadID != "" {
		uploadID = username + fmt.Sprintf("%x", time.Now().UnixNano())
	} else {
		chunks, err := redis.Values(redisConn.Do("HGETALL", ChunkKeyPrefix + uploadID))
		if err != nil {
			w.Write(util.NewRespMsg(-1, "Upload part failed", err.Error()).JSONBytes())
			return
		}
		for i := 0; i < len(chunks); i += 2 {
			key := string(chunks[i].([]byte))
			val := string(chunks[i + 1].([]byte))
			if strings.HasPrefix(key, "chkidx_") && val == "1" {
				chunkIdx, _ := strconv.Atoi(key[7:])
				chunksExist = append(chunksExist, chunkIdx)
			}
		}
	}

	// 5. Initialize multipart upload information
	upInfo := MultipartUploadInfo{
		FileHash: filehash,
		FileSize: filesize,
		UploadID: uploadID,
		ChunkSize: 5 * 1024 * 1024,
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
		ChunkExists: chunksExist,
	}

	// 6. Write multipart information into redis
	if len(upInfo.ChunkExists) <= 0 {
		hkey := "MP_" + upInfo.UploadID
		redisConn.Do("HSET", hkey, "chunkcount", upInfo.ChunkCount)
		redisConn.Do("HSET", hkey, "filehash", upInfo.FileHash)
		redisConn.Do("HSET", hkey, "filesize", upInfo.FileSize)
		redisConn.Do("EXPIRE", hkey, 43200)
		redisConn.Do("SET", HashUpIDKeyPrefix + filehash, upInfo.UploadID, "EX", 43200)
	}

	// 7. Return response back to client
	w.Write(util.NewRespMsg(0, "OK", upInfo).JSONBytes())
}

func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request params
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

	// 2. Get redis pool connection
	redisConn := redisPool.RedisPool().Get()
	defer redisConn.Close()

	// 3. Get a file descriptor to store chunk
	fpath := "/data/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		fmt.Println(err.Error())
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer fd.Close()

	buf := make([]byte, 1024 * 1024)
	for {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	// 4. Update redis cache with chunk status
	redisConn.Do("HSET", "MP_" + uploadID, "chkidx_" + chunkIndex, 1)

	// 5. Return response back to client
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request params
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")

	// 2. Get redis pool connection
	redisConn := redisPool.RedisPool().Get()
	defer redisConn.Close()

	// 3. Check in redis that all chunks are uploaded
	data, err := redis.Values(redisConn.Do("HGETALL", "MP_" + uploadID))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Complete upload failed", nil).JSONBytes())
		return
	}
	totalCount := 0
	chunkCount := 0
	for i := 1; i < len(data); i += 2 {
		key := string(data[i].([]byte))
		val := string(data[i].([]byte))
		if key == "chunkcount" {
			totalCount, _ = strconv.Atoi(val)
		} else if strings.HasPrefix(key, "chkidx_") && val == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(-2, "invalid request", nil).JSONBytes())
		return
	}

	// 4. combine chunks
	if mergeSuc := util.MergeChuncksByShell(ChunkDir+uploadID, MergeDir+filehash, filehash); !mergeSuc {
		w.Write(util.NewRespMsg(-3, "Complete upload failed", nil).JSONBytes())
		return
	}

	// 5. Update file table and user file table
	fsize, _ := strconv.Atoi(filesize)
	dblayer.OnFileUploadFinished(filehash, filename, int64(fsize), "")
	dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(fsize))

	// 6. Response back to client
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}