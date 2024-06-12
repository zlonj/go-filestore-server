package handler

import (
	"context"
	"encoding/json"
	"filestore-server/config"
	dblayer "filestore-server/db"
	"filestore-server/meta"
	appS3 "filestore-server/store/s3"
	"filestore-server/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Handles file upload
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Return upload html page
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internal server error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == http.MethodPost {
		// Accept file stream and save to local directory
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("Failed to get file data, err: %s\n", err.Error())
			return
		}
		defer file.Close()

		newFile, err := os.Create("/tmp/" + head.Filename)
		if err != nil {
			fmt.Printf("Failed to create file, err: %s\n", err.Error())
			return
		}
		defer newFile.Close()

		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "/tmp/" + head.Filename,
			UploadAt: time.Now().Format(time.DateTime),
		}

		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("Failed to save data into file, err %s\n", err.Error())
			return
		}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)

		// Store file to S3
		s3Bucket := config.S3_BUCKET
		s3Key := fileMeta.FileSha1
		_, err = appS3.Client().PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: &s3Bucket,
			Key: &s3Key,
		})
		if err != nil {
			fmt.Println("Upload S3 err: " + err.Error())
			w.Write([]byte("Upload failed"))
			return
		}
		fileMeta.Location = s3Key

		// Update/insert file metadata record into mysql
		_ = meta.UpdateFileMetadataDB(fileMeta)

		// TODO: Update user file table
		r.ParseForm()
		username := r.Form.Get("username")
		suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
		if suc {
			fmt.Printf("File uploaded with hash: %s", fileMeta.FileSha1)
			http.Redirect(w, r, "/file/upload/success", http.StatusFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Upload Failed."))
		}
	}
}

// Handles successful file upload
func UploadSuccessHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "File upload successful!")
}

// Handles get file metadata
func GetFileMetadataHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fileHash := r.Form["filehash"][0]
	fileMeta, err := meta.GetFileMetadataDB(fileHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(fileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	limitCount, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	userFiles, err := dblayer.QueryUserFileMetas(username, limitCount)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fileHash := r.Form.Get("filehash")
	fileMeta := meta.GetFileMeta(fileHash)

	// Retrieve object from S3
	s3Bucket := config.S3_BUCKET
	s3Key := fileMeta.FileSha1
	resp, err := appS3.Client().GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &s3Bucket,
		Key: &s3Key,
	})
	if err != nil {
		fmt.Println("Failed to get S3 object, err: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to download file"))
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileMeta.FileName+"\"")
	w.Write(data)
}

func FileMetadataUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	operationType := r.Form.Get("op")
	fileHash := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if operationType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Rename and update file metadata
	fileMeta := meta.GetFileMeta(fileHash)
	fileMeta.FileName = newFileName
	meta.UpdateFileMeta(fileMeta)

	data, err := json.Marshal(fileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// Deletes a file by file hash
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fileHash := r.Form.Get("filehash")

	// Remove file from file system
	fileMeta := meta.GetFileMeta(fileHash)
	err := os.Remove(fileMeta.Location)
	if err != nil {
		fmt.Printf("Failed to delete file %s at location %s\n", fileMeta.FileName, fileMeta.Location)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	meta.RemoveFileMeta(fileHash)
	w.WriteHeader(http.StatusOK)
}

func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// 1. Parse request params
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))

	// 2. Look up file hash from table_file
	fileMeta, err := meta.GetFileMetadataDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3. If no record, return failure
	if fileMeta.FileSha1 == "" {
		resp := util.RespMsg{
			Code: -1,
			Msg: "Fast upload failed, call normal upload API",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 4. Otherwise, insert user file record into table user file
	suc := dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg: "Fast upload successful",
		}
		w.Write(resp.JSONBytes())
		return
	}

	resp := util.RespMsg{
		Code: -2,
		Msg: "Fast upload failed, please retry later",
	}
	w.Write(resp.JSONBytes())
}