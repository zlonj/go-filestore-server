package meta

import "sort"

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