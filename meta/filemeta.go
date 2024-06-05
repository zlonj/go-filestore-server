package meta

import (
	"sort"
	mydb "filestore-server/db"
)

// Metadata for files
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)
}

// Add/update a file's metadata
func UpdateFileMeta(fmeta FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}

// Get file metadata by file's hash value
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

// Add/update file metadata to mysql DB
func UpdateFileMetadataDB(fileMeta FileMeta) bool {
	return mydb.OnFileUploadFinished(
		fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize, fileMeta.Location)
}

// Get file metadata from DB
func GetFileMetadataDB(fileHash string) (FileMeta, error) {
	tableFile, err := mydb.GetFileMeta(fileHash)
	if err != nil {
		return FileMeta{}, err
	}
	fileMeta := FileMeta{
		FileSha1: tableFile.FileHash,
		FileName: tableFile.FileName.String,
		FileSize: tableFile.FileSize.Int64,
		Location: tableFile.FileAddr.String,
	}
	return fileMeta, nil
}

// Get last `count` file metadata by upload time
func GetLastFileMetas(count int) []FileMeta {
	var fileMetaArray []FileMeta
	for _, meta := range fileMetas {
		fileMetaArray = append(fileMetaArray, meta)
	}

	sort.Sort(ByUploadTime(fileMetaArray))
	if count > len(fileMetaArray) {
		return fileMetaArray
	}
	return fileMetaArray[0:count]
}

// Delete a file metadata
func RemoveFileMeta(fileHash string) {
	delete(fileMetas, fileHash)
}