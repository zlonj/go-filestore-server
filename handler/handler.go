package handler

import (
	"encoding/json"
	"filestore-server/meta"
	"filestore-server/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Handles file upload
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Return upload html page
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internal server error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
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
		 meta.UpdateFileMeta(fileMeta)

		http.Redirect(w, r, "/file/upload/success", http.StatusFound)
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
	fileMeta := meta.GetFileMeta(fileHash)
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
	fileMetas := meta.GetLastFileMetas(limitCount)
	data, err := json.Marshal(fileMetas)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}