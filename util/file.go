package util

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	MergeFileCMD = `
	#!/bin/bash
	chunkDir=$1
	mergePath=$2
		
	if [ ! -f $mergePath ]; then
			echo "$mergePath not exist"
	else
			rm -f $mergePath
	fi
	
	for chunk in $(ls $chunkDir | sort -n)
	do
			cat $chunkDir/${chunk} >> ${mergePath}
	done
	`

	FileSha1CMD = `
	#!/bin/bash
	sha1sum $1 | awk '{print $1}'
	`

	FileSizeCMD = `
	#!/bin/bash
	ls -l $1 | awk '{print $5}'
	`

	FileChunksDelCMD = `
	#!/bin/bash
	chunkDir="/data/chunks/"
	targetDir=$1
	if [[ $targetDir =~ $chunkDir ]] && [[ $targetDir != $chunkDir ]]; then 
	  rm -rf $targetDir
	fi
	`
)

func RemovePathByShell(destPath string) bool {
	cmdStr := strings.Replace(FileChunksDelCMD, "$1", destPath, 1)
	delCmd := exec.Command("bash", "-c", cmdStr)
	if _, err := delCmd.Output(); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func ComputeFileSizeByShell(destPath string) (int, error) {
	cmdStr := strings.Replace(FileSizeCMD, "$1", destPath, 1)
	fSizeCmd := exec.Command("bash", "-c", cmdStr)
	if fSizeStr, err := fSizeCmd.Output(); err != nil {
		fmt.Println(err)
		return -1, err
	} else {
		reg := regexp.MustCompile("\\s+")
		fSize, err := strconv.Atoi(reg.ReplaceAllString(string(fSizeStr), ""))
		if err != nil {
			fmt.Println(err)
			return -1, err
		}
		return fSize, nil
	}
}

func ComputeSha1ByShell(destPath string) (string, error) {
	cmdStr := strings.Replace(FileSha1CMD, "$1", destPath, 1)
	hashCmd := exec.Command("bash", "-c", cmdStr)
	if filehash, err := hashCmd.Output(); err != nil {
		fmt.Println(err)
		return "", err
	} else {
		reg := regexp.MustCompile("\\s+")
		return reg.ReplaceAllString(string(filehash), ""), nil
	}
}

func MergeChuncksByShell(chunkDir string, destPath string, fileSha1 string) bool {
	cmdStr := strings.Replace(MergeFileCMD, "$1", chunkDir, 1)
	cmdStr = strings.Replace(cmdStr, "$2", destPath, 1)
	mergeCmd := exec.Command("bash", "-c", cmdStr)
	if _, err := mergeCmd.Output(); err != nil {
		fmt.Println(err)
		return false
	}

	if filehash, err := ComputeSha1ByShell(destPath); err != nil {
		fmt.Println(err)
		return false
	} else if string(filehash) != fileSha1 {
		fmt.Println(filehash + " " + fileSha1)
		return false
	} else {
		fmt.Println("check sha1: " + destPath + " " + filehash + " " + fileSha1)
	}

	return true
}
