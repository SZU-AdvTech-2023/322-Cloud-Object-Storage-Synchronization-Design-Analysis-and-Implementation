package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type FileStatus struct {
	FileName  string //File name
	FileMtime string // Modify time
	FileId    string //the hash value of file
}

func (fs *FileStatus) init(fileName string, fileMtime string, fileId string) {
	fs.FileName = fileName
	fs.FileMtime = fileMtime
	fs.FileId = fileId
}

func FSeq(fs, other FileStatus) bool {
	return fs.FileName == other.FileName
}

func FSlt(fs, other FileStatus) bool {
	return fs.FileName < other.FileName
}

func (fs *FileStatus) toString() string {
	return "FileStatus: filename=" + fs.FileName + ", fileID=" + fs.FileId
}

type DirectoryStatus struct {
	DirectoryName  string
	DirectoryMtime string            // Modify time
	DirectoryId    string            // Hash value
	ChildrenFile   []FileStatus      // a slice for subfile
	ChildrenDirect []DirectoryStatus // a slice for subdir
}

func (ds *DirectoryStatus) init(directoryName string, directoryMtime string, directoryId string, childrenFile []FileStatus, childrenDirect []DirectoryStatus) {
	ds.DirectoryName = directoryName
	ds.DirectoryMtime = directoryMtime
	ds.DirectoryId = directoryId
	ds.ChildrenFile = childrenFile
	ds.ChildrenDirect = childrenDirect
}

func DSeq(ds, other DirectoryStatus) bool {
	return ds.DirectoryName == other.DirectoryName
}

func DSlt(ds, other DirectoryStatus) bool {
	return ds.DirectoryName < other.DirectoryName
}

func (ds *DirectoryStatus) toString() string {
	return "DirectoryStatus: directoryName=" + ds.DirectoryName + ", directoryId=" + ds.DirectoryId
}

func (ds *DirectoryStatus) insertSubFile(fs FileStatus) {
	ds.ChildrenFile = append(ds.ChildrenFile, fs)
}

func (ds *DirectoryStatus) insertSubDirectory(other DirectoryStatus) {
	ds.ChildrenDirect = append(ds.ChildrenDirect, other)
}

func (ds *DirectoryStatus) getSameSubFile(other FileStatus) *FileStatus {
	for _, subFile := range ds.ChildrenFile {
		if FSeq(subFile, other) {
			return &subFile
		}
	}
	return nil
}

func (ds *DirectoryStatus) getSameSubDirectory(other DirectoryStatus) *DirectoryStatus {
	for _, subDirectory := range ds.ChildrenDirect {
		if DSeq(subDirectory, other) {
			return &subDirectory
		}
	}
	return nil
}

func getLocalSubFiles(localPath string) []FileStatus {
	//get the subfiles of the localPath
	if localPath[len(localPath)-1] != '/' {
		localPath += "/"
	}
	//get the subfiles
	var subFiles []FileStatus
	files, err := ioutil.ReadDir(localPath)
	if err != nil {
		fmt.Println("无法读取目录:", err)
		os.Exit(1)
	}

	for _, file := range files {
		if !file.IsDir() {
			fmt.Println(file.Name())
			var tempFS FileStatus
			tempFS.init(file.Name(), strconv.Itoa(int(file.ModTime().Unix())), getLocalFileHash(localPath+file.Name()))
			subFiles = append(subFiles, tempFS)
		}
	}
	sort.Slice(subFiles, func(i, j int) bool {
		return FSlt(subFiles[i], subFiles[j])
	})
	return subFiles
}

func getLocalSubDirects(localPath string) []DirectoryStatus {
	//get the subfiles of the localPath
	if localPath[len(localPath)-1] != '/' {
		localPath += "/"
	}
	//get the subDirectories
	var subDirectories []DirectoryStatus
	directories, err := ioutil.ReadDir(localPath)
	if err != nil {
		fmt.Println("无法读取目录:", err)
		os.Exit(1)
	}

	for _, directory := range directories {
		if directory.IsDir() {
			fmt.Println(directory.Name())
			var tempDS DirectoryStatus
			tempDS.init(directory.Name()+"/", strconv.Itoa(int(directory.ModTime().Unix())), getBufferHash([]byte("")), make([]FileStatus, 0), make([]DirectoryStatus, 0))
			subDirectories = append(subDirectories, tempDS)
		}
	}
	sort.Slice(subDirectories, func(i, j int) bool {
		return DSlt(subDirectories[i], subDirectories[j])
	})
	return subDirectories
}

func buildLocalMetaTree(localPath string) DirectoryStatus {
	//Build the local meta tree, the localPath is the path of the root
	if localPath[len(localPath)-1] != '/' {
		localPath += "/"
	}
	//initialize the root
	var root DirectoryStatus
	DirectoryInfo, err := os.Stat(localPath)
	if err != nil {
		panic(err)
	}
	root.init(DirectoryInfo.Name()+"/", strconv.Itoa(int(DirectoryInfo.ModTime().Unix())), getBufferHash([]byte("")), make([]FileStatus, 0), make([]DirectoryStatus, 0))
	subFiles := getLocalSubFiles(localPath)
	subDirects := getLocalSubDirects(localPath)
	//First handle the subDirects
	if len(subDirects) != 0 {
		for _, subDirect := range subDirects {
			nextLocalPath := localPath + subDirect.DirectoryName
			root.insertSubDirectory(buildLocalMetaTree(nextLocalPath))
		}
	}
	//Then handle the subFiles
	if len(subFiles) != 0 {
		for _, subFile := range subFiles {
			root.insertSubFile(subFile)
			hashMapLocal[subFile.FileId] = append(hashMapLocal[subFile.FileId], localPath+subFile.FileName)
		}
	}
	return root
}

func getCloudSubFiles(cloudPath string) []FileStatus {
	var subFiles []FileStatus
	for _, subFileName := range cfs.listFiles(cloudPath) {
		if subFileName[len(subFileName)-1] != '/' {
			tempPath := cloudPath + subFileName
			stat := cfs.statFile(tempPath)
			var tempFile FileStatus
			tempFile.init(subFileName, stat["mtime"], stat["hash"])
			subFiles = append(subFiles, tempFile)
		}
	}
	sort.Slice(subFiles, func(i, j int) bool {
		return FSlt(subFiles[i], subFiles[j])
	})
	return subFiles
}

func getCloudSubDirects(cloudPath string) []DirectoryStatus {
	var subDirects []DirectoryStatus
	for _, subDirectName := range cfs.listFiles(cloudPath) {
		if subDirectName[len(subDirectName)-1] == '/' {
			tempPath := cloudPath + subDirectName
			stat := cfs.statFile(tempPath)
			var tempDirect DirectoryStatus
			tempDirect.init(subDirectName, stat["mtime"], stat["hash"], make([]FileStatus, 0), make([]DirectoryStatus, 0))
			subDirects = append(subDirects, tempDirect)
		}
	}
	sort.Slice(subDirects, func(i, j int) bool {
		return DSlt(subDirects[i], subDirects[j])
	})
	return subDirects
}

func buildCloudMetaTree(cloudPath string) DirectoryStatus {
	//build the cloud meta tree, the cloudPath is the path of the root
	if cloudPath[len(cloudPath)-1] != '/' {
		cloudPath += "/"
	}
	var root DirectoryStatus
	//initialize the root
	stat := cfs.statFile(cloudPath)
	root.init(filepath.Base(cloudPath)+"/", stat["mtime"], stat["hash"], make([]FileStatus, 0), make([]DirectoryStatus, 0))
	subFiles := getCloudSubFiles(cloudPath)
	subDirects := getCloudSubDirects(cloudPath)
	//handle the subDirects
	if len(subDirects) != 0 {
		for _, subDirect := range subDirects {
			nextCloudPath := cloudPath + subDirect.DirectoryName
			root.insertSubDirectory(buildCloudMetaTree(nextCloudPath))
		}
	}
	//handle the subFiles
	if len(subFiles) != 0 {
		for _, subFile := range subFiles {
			root.insertSubFile(subFile)
			hashMapCloud[subFile.FileId] = append(hashMapCloud[subFile.FileId], cloudPath+subFile.FileName)
		}
	}
	return root
}
