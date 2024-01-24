package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

var cfs CloudFileSystem
var historyPath string
var localPath string
var cloudPath string

var tasks synchronizeEventEmitter
var cloudMetaTree DirectoryStatus
var localMetaTree DirectoryStatus
var historyMetaTree DirectoryStatus

var isEnd bool

func initSynchronization(cloudFileSystem CloudFileSystem) {
	cfs = cloudFileSystem
	historyPath = cloudFileSystem.config["history_path"]
	localPath = cloudFileSystem.config["local_path"]
	cloudPath = cloudFileSystem.config["cloud_path"]

	tasks.init()
	isEnd = false
	var observer synchronizeEventHandler
	observer.init(cfs)
	tasks.register(observer)
	err := os.Chdir(localPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	hashMapLocal = make(map[string][]string)
	hashMapCloud = make(map[string][]string)
	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, syscall.SIGINT, syscall.SIGTERM) // When you enter the 'Ctrl+C', the symbol isEnd will be set to true
	go func() {
		<-interruptSignal
		isEnd = true
	}()
}

func start() {
	//Start the synchronization
	//you can customize
	fmt.Println("The Cloud Synchronization System is starting.")
	fmt.Println("You can enter 'Ctrl+C' to stop the system.")
	isEnd = false
	initialize()
	for !isEnd {
		synchronzie()
		time.Sleep(30)
	}
	fmt.Println("Stop synchronizing")
	fmt.Println("The system has been stopped.")
}

func initialize() {
	//buildTree
	//cloudMetaTree = buildCloudMetaTree(cloudPath)
	//localMetaTree = buildCloudMetaTree(localPath)
	//Get the historyTree
	historyTreePath := historyPath + "history_tree.gob"
	stat, err := os.Stat(historyTreePath)
	if os.IsNotExist(err) { // the history_tree has not built yet
		file, err1 := os.Create(historyTreePath)
		if err1 != nil {
			fmt.Println(err1)
			os.Exit(-1)
		}
		file.Close()
	}
	err = os.Chmod(historyTreePath, 0777)
	if err != nil {
		panic(err)
	}
	if stat.Size() > 0 {
		// the history_tree has already existed and not none, open it
		file, err1 := os.Open(historyPath + "history_tree.gob")
		if err1 != nil {
			panic(err1)
		}
		decoder := gob.NewDecoder(file)
		err1 = decoder.Decode(&historyMetaTree)
		if err1 != nil {
			panic(err1)
		}
	} else {
		// the history_tree is none
		historyMetaTree.DirectoryName = filepath.Base(cloudPath) + "/"
	}
}

func synchronzie() {

	//build the cloud current meta tree
	cloudMetaTree = buildCloudMetaTree(cloudPath)

	//run pull algorithm
	algorithmPull(cloudMetaTree, historyMetaTree, cloudPath, localPath)

	historyMetaTree = cloudMetaTree

	//build the local current meta tree
	localMetaTree = buildLocalMetaTree(localPath)

	//run push algorithm
	algorithmPush(localMetaTree, historyMetaTree, cloudPath, localPath)

	historyMetaTree = localMetaTree
	saveHistory()
	err := os.RemoveAll(home + "/tempCloudFile/")
	if err != nil {
		panic(err)
	}
}

func saveHistory() {
	file, err := os.OpenFile(historyPath+"history_tree.gob", os.O_RDWR, 0777)
	if err != nil {
		panic(err)
	}
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(historyMetaTree)
	if err != nil {
		panic(err)
	}
}

func algorithmPush(local DirectoryStatus, historyTree DirectoryStatus, cPath string, lPath string) {
	//new
	//handle the subDirectory
	for _, subDirectory := range local.ChildrenDirect {
		DirectoryName := subDirectory.DirectoryName
		nextLocalPath := lPath + DirectoryName
		nextCloudPath := cPath + DirectoryName
		nextHistory := historyTree.getSameSubDirectory(subDirectory)
		if nextHistory != nil {
			algorithmPush(subDirectory, *nextHistory, nextCloudPath, nextLocalPath)
		}
		if nextHistory == nil {
			args := make([]string, 0)
			args = append(args, nextLocalPath)
			args = append(args, nextCloudPath)
			tasks.setData(CREATE_CLOUD_FOLDER, args, map[string]string{})
		}
	}
	for _, subFile := range local.ChildrenFile {
		fileName := subFile.FileName
		nextLocalPath := lPath + fileName
		nextCloudPath := cPath + fileName
		nextHistory := historyTree.getSameSubFile(subFile)
		if nextHistory == nil {
			args := make([]string, 0)
			args = append(args, nextLocalPath)
			args = append(args, nextCloudPath)
			tasks.setData(UPLOAD_FILE, args, map[string]string{"file_id": subFile.FileId})
		}
		if nextHistory != nil {
			historyMtime, _ := strconv.Atoi(nextHistory.FileMtime)
			localMtime, _ := strconv.Atoi(subFile.FileMtime)
			if nextHistory != nil && historyMtime < localMtime && nextHistory.FileId != subFile.FileId {
				args := make([]string, 0)
				args = append(args, nextLocalPath)
				args = append(args, nextCloudPath)
				tasks.setData(UPDATE_CLOUD_FILE, args, map[string]string{"file_id": subFile.FileId})
			}
		}
	}
	//delete
	if historyTree.DirectoryName == "" {
		return
	}
	//handle the subDirectory
	for _, nextHistory := range historyTree.ChildrenDirect {
		directoryName := nextHistory.DirectoryName
		//nextLocalPath := localPath + directoryName
		nextCloudPath := cPath + directoryName
		nextLocal := local.getSameSubDirectory(nextHistory)
		if nextLocal == nil {
			args := make([]string, 0)
			args = append(args, nextCloudPath)
			tasks.setData(DELETE_CLOUD_FOLDER, args, map[string]string{})
		}
	}
	for _, nextHistory := range historyTree.ChildrenFile {
		fileName := nextHistory.FileName
		//nextLocalPath := localPath+fileName
		nextCloudPath := cPath + fileName
		nextLocal := local.getSameSubFile(nextHistory)
		if nextLocal == nil {
			args := make([]string, 0)
			args = append(args, nextCloudPath)
			tasks.setData(DELETE_CLOUD_FILE, args, map[string]string{})
		}
	}
}

func algorithmPull(cloud DirectoryStatus, historyTree DirectoryStatus, cPath string, lPath string) {
	//new
	//handle the subDirectory
	for _, subDirectory := range cloud.ChildrenDirect {
		DirectoryName := subDirectory.DirectoryName
		nextLocalPath := lPath + DirectoryName
		nextCloudPath := cPath + DirectoryName
		nextHistory := historyTree.getSameSubDirectory(subDirectory)
		if nextHistory != nil {
			algorithmPull(subDirectory, *nextHistory, nextCloudPath, nextLocalPath)
		}
		if nextHistory == nil {
			args := make([]string, 0)
			args = append(args, nextCloudPath)
			args = append(args, nextLocalPath)
			tasks.setData(CREATE_LOCAL_FOLDER, args, map[string]string{})
		}
	}
	for _, subFile := range cloud.ChildrenFile {
		fileName := subFile.FileName
		nextLocalPath := lPath + fileName
		nextCloudPath := cPath + fileName
		nextHistory := historyTree.getSameSubFile(subFile)
		if nextHistory == nil {
			args := make([]string, 0)
			args = append(args, nextCloudPath)
			args = append(args, nextLocalPath)
			tasks.setData(DOWNLOAD_FILE, args, map[string]string{"file_id": subFile.FileId})
		}
		if nextHistory != nil {
			historyMtime, _ := strconv.Atoi(nextHistory.FileMtime)
			localMtime, _ := strconv.Atoi(subFile.FileMtime)
			if nextHistory != nil && historyMtime < localMtime && nextHistory.FileId != subFile.FileId {
				args := make([]string, 0)
				args = append(args, nextCloudPath)
				args = append(args, nextLocalPath)
				tasks.setData(UPDATE_LOCAL_FILE, args, map[string]string{"file_id": subFile.FileId})
			}
		}
	}
	//delete
	if historyTree.DirectoryName == "" {
		return
	}
	//handle the subDirectory
	for _, nextHistory := range historyTree.ChildrenDirect {
		directoryName := nextHistory.DirectoryName
		nextLocalPath := localPath + directoryName
		//nextCloudPath := cPath + directoryName
		nextLocal := cloud.getSameSubDirectory(nextHistory)
		if nextLocal == nil {
			args := make([]string, 0)
			args = append(args, nextLocalPath)
			tasks.setData(DELETE_LOCAL_FOLDER, args, map[string]string{})
		}
	}
	for _, nextHistory := range historyTree.ChildrenFile {
		fileName := nextHistory.FileName
		nextLocalPath := localPath + fileName
		//nextCloudPath := cPath + fileName
		nextLocal := cloud.getSameSubFile(nextHistory)
		if nextLocal == nil {
			args := make([]string, 0)
			args = append(args, nextLocalPath)
			tasks.setData(DELETE_LOCAL_FILE, args, map[string]string{})
		}
	}
}
