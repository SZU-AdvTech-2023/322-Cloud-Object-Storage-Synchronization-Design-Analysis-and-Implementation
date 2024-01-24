package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

type synchronizeEventHandler struct {
	cfs      CloudFileSystem
	fromPath string
	toPath   string
	kwargs   map[string]string
}

func (observer *synchronizeEventHandler) init(cfs CloudFileSystem) {
	observer.cfs = cfs
	observer.fromPath = ""
	observer.toPath = ""
	observer.kwargs = map[string]string{}
}

func SEHeq(observer1, observer2 synchronizeEventHandler) bool {
	return observer1.cfs.serviceProvider == observer2.cfs.serviceProvider &&
		observer1.cfs.config["history_path"] == observer2.cfs.config["history_path"] &&
		observer1.cfs.config["local_path"] == observer2.cfs.config["local_path"] &&
		observer1.cfs.config["cloud_path"] == observer2.cfs.config["cloud_path"]
}

func (observer *synchronizeEventHandler) createCloudFolder(fromPath, toPath string) {
	if fromPath == "" {
		fromPath = observer.fromPath
	}
	if toPath == "" {
		toPath = observer.toPath
	}
	//Check if the folder exists
	_, err := os.Stat(fromPath)
	if os.IsNotExist(err) {
		fmt.Println("The local folder " + fromPath + " does not exist")
		return
	}
	if observer.cfs.statFile(toPath) != nil {
		fmt.Println("The cloud folder " + toPath + " does exist")
		return
	}
	if fromPath[len(fromPath)-1] != '/' {
		fromPath += "/"
	}
	if toPath[len(toPath)-1] != '/' {
		toPath += "/"
	}
	//Create the folder
	observer.cfs.createFolder(toPath)
	files, err1 := ioutil.ReadDir(fromPath)
	if err1 != nil {
		panic(err1)
	}
	for _, file := range files {
		nextFromPath := fromPath + file.Name()
		if file.IsDir() {
			observer.createCloudFolder(nextFromPath+"/", toPath+file.Name()+"/")
		} else {
			observer.upload(nextFromPath, toPath+file.Name())
		}
	}
}

func (observer *synchronizeEventHandler) createLocalFolder(fromPath, toPath string) {
	if fromPath == "" {
		fromPath = observer.fromPath
	}
	if toPath == "" {
		toPath = observer.toPath
	}
	if fromPath[len(fromPath)-1] != '/' {
		fromPath += "/"
	}
	if toPath[len(toPath)-1] != '/' {
		toPath += "/"
	}
	//Check if the folder has already existed
	if observer.cfs.statFile(fromPath) == nil {
		fmt.Println("The cloud folder " + fromPath + " does not exist")
		return
	}
	_, err := os.Stat(toPath)
	if err == nil || os.IsExist(err) {
		fmt.Println("The local folder " + toPath + " does exist")
		return
	}
	err = os.Mkdir(toPath, 0755)
	if err != nil {
		fmt.Println("err")
		return
	}
	for _, fileName := range observer.cfs.listFiles(fromPath) {
		if fileName[len(fileName)-1] == '/' {
			observer.createLocalFolder(fromPath+fileName, toPath+fileName)
		} else {
			observer.download(fromPath+fileName, toPath+fileName)
		}
	}
}

func (observer *synchronizeEventHandler) upload(fromPath, toPath string) {
	if fromPath == "" {
		fromPath = observer.fromPath
	}
	if toPath == "" {
		toPath = observer.toPath
	}
	//Check the file
	_, err := os.Stat(fromPath)
	if os.IsNotExist(err) {
		fmt.Println("Local file doesn't exist: " + fromPath)
		return
	}
	if observer.cfs.statFile(toPath) != nil {
		fmt.Println("The file has already existed on the cloud: " + toPath)
	}

	if observer.kwargs["file_id"] != "" {
		// If the file has already exist on the cloud with a different path
		records := hashMapCloud[observer.kwargs["file_id"]]
		if len(records) > 0 {
			for _, record := range records {
				observer.cfs.copy(record, toPath)
				return
			}
		}
	}
	//upload
	observer.cfs.upload(toPath, fromPath)
}

func (observer *synchronizeEventHandler) download(fromPath, toPath string) {
	if fromPath == "" {
		fromPath = observer.fromPath
	}
	if toPath == "" {
		toPath = observer.toPath
	}

	// Check the file
	if observer.cfs.statFile(fromPath) == nil {
		fmt.Println("The file doesn't exist on the cloud: " + fromPath)
		return
	}
	_, err := os.Stat(toPath)
	if err == nil || os.IsExist(err) {
		fmt.Println("The file has existed in local: " + toPath)
		return
	}

	if observer.kwargs["file_id"] != "" {
		//The file has existed in local with a different path
		records := hashMapLocal[observer.kwargs["file_id"]]
		if len(records) > 0 {
			for _, record := range records {
				err1 := os.Rename(record, toPath)
				if err1 != nil {
					continue
				}
				return
			}
		}
	}
	//download
	observer.cfs.download(fromPath, toPath)
}

func (observer *synchronizeEventHandler) deleteCloudFile() {
	var cloudPath = observer.fromPath
	if observer.cfs.statFile(cloudPath) != nil {
		fmt.Println("The file doesn't exist on the cloud: " + cloudPath)
		return
	}

	//delete
	observer.cfs.delete(cloudPath)
}

func (observer *synchronizeEventHandler) deleteLocalFile() {
	var localPath = observer.fromPath
	_, err := os.Stat(localPath)
	if os.IsNotExist(err) {
		fmt.Println("The file doesn't exist in local: " + localPath)
		return
	}

	//delete
	err = os.Remove(localPath)
	if err != nil {
		panic(err)
	}
}

func (observer *synchronizeEventHandler) deleteCloudFolder(cloudPath string) {
	if cloudPath == "" {
		cloudPath = observer.fromPath
	}
	if cloudPath[len(cloudPath)-1] != '/' {
		cloudPath += "/"
	}

	// Check the folder
	if observer.cfs.statFile(cloudPath) == nil {
		fmt.Println("The folder doesn't exist on the cloud: " + cloudPath)
		return
	}

	//delete
	for _, fileName := range observer.cfs.listFiles(cloudPath) {
		if fileName[len(fileName)-1] == '/' {
			observer.deleteCloudFolder(cloudPath + fileName)
		} else {
			observer.cfs.delete(cloudPath + fileName)
		}
	}
	observer.cfs.delete(cloudPath)
}

func (observer *synchronizeEventHandler) deleteLocalFolder(localPath string) {
	if localPath == "" {
		localPath = observer.fromPath
	}
	if localPath[len(localPath)-1] != '/' {
		localPath += "/"
	}

	//Check the folder
	_, err := os.Stat(localPath)
	if os.IsNotExist(err) {
		fmt.Println("The folder doesn't exist in local: " + localPath)
		return
	}

	//delete
	files, err1 := ioutil.ReadDir(localPath)
	if err1 != nil {
		panic(err1)
	}
	for _, file := range files {
		if file.IsDir() {
			observer.deleteLocalFolder(localPath + file.Name() + "/")
		} else {
			err2 := os.Remove(localPath + file.Name())
			if err2 != nil {
				continue
			}
		}
	}
	err1 = os.Remove(localPath)
	//err1 = os.RemoveAll(localPath)
	if err1 != nil {
		panic(err1)
	}
}

func (observer *synchronizeEventHandler) updateCloudFile() {
	fromPath := observer.fromPath
	toPath := observer.toPath
	if observer.kwargs["file_id"] != "" {
		records := hashMapCloud[observer.kwargs["file_id"]]
		if len(records) > 0 {
			for _, record := range records {
				observer.cfs.copy(record, toPath)
				return
			}
		}
	}
	observer.cfs.update(fromPath, toPath)
}

func (observer *synchronizeEventHandler) updateLocalFile() {
	fromPath := observer.fromPath
	toPath := observer.toPath
	if observer.kwargs["file_id"] != "" {
		records := hashMapLocal[observer.kwargs["file_id"]]
		if len(records) > 0 {
			for _, record := range records {
				err := os.Rename(record, toPath)
				if err != nil {
					continue
				}
				return
			}
		}
	}
	observer.cfs.download(fromPath, toPath)
}

func (observer *synchronizeEventHandler) renameCloudFile() {
	fromPath := observer.fromPath
	toPath := observer.toPath

	//Check the file
	if observer.cfs.statFile(fromPath) == nil {
		fmt.Println("The file doesn't exist on the cloud: " + fromPath)
		return
	}
	if observer.cfs.statFile(toPath) != nil {
		fmt.Println("The file has already existed on the cloud: " + toPath)
		return
	}

	//rename
	observer.cfs.rename(fromPath, toPath)
}

func (observer *synchronizeEventHandler) renameLocalFile() {
	fromPath := observer.fromPath
	toPath := observer.toPath

	//Check the file
	_, err1 := os.Stat(fromPath)
	_, err2 := os.Stat(toPath)
	if os.IsNotExist(err1) {
		fmt.Println("The file doesn't exist in local: " + fromPath)
		return
	}
	if err2 == nil || os.IsExist(err2) {
		fmt.Println("The file has already existed in local: " + toPath)
		return
	}

	//rename
	err := os.Rename(fromPath, toPath)
	if err != nil {
		panic(err)
	}
}

func (observer *synchronizeEventHandler) renameCloudFolder() {
	fromPath := observer.fromPath
	toPath := observer.toPath
	if fromPath[len(fromPath)-1] != '/' {
		fromPath += "/"
	}
	if toPath[len(toPath)-1] != '/' {
		toPath += "/"
	}
	//Check the folder
	if observer.cfs.statFile(fromPath) == nil {
		fmt.Println("The folder doesn't exist on the cloud: " + fromPath)
		return
	}
	if observer.cfs.statFile(toPath) != nil {
		fmt.Println("The folder has already exist on the cloud: " + toPath)
		return
	}

	observer.cfs.rename(fromPath, toPath)
}

func (observer *synchronizeEventHandler) renameLocalFolder() {
	fromPath := observer.fromPath
	toPath := observer.toPath
	if fromPath[len(fromPath)-1] != '/' {
		fromPath += "/"
	}
	if toPath[len(toPath)-1] != '/' {
		toPath += "/"
	}

	//Check the file
	_, err1 := os.Stat(fromPath)
	_, err2 := os.Stat(toPath)
	if os.IsNotExist(err1) {
		fmt.Println("The folder doesn't exist in local: " + fromPath)
		return
	}
	if err2 == nil || os.IsExist(err2) {
		fmt.Println("The folder has already existed in local: " + toPath)
		return
	}

	err := os.Rename(fromPath, toPath)
	if err != nil {
		return
	}
}

func (observer *synchronizeEventHandler) update(observable synchronizeEventEmitter) {
	taskIndex := observable.taskIndex
	observer.fromPath = observable.fromPath
	observer.toPath = observable.toPath
	observer.kwargs = observable.kwargs

	switch taskIndex {
	case CREATE_CLOUD_FOLDER:
		observer.createCloudFolder("", "")
		break
	case CREATE_LOCAL_FOLDER:
		observer.createLocalFolder("", "")
		break
	case UPLOAD_FILE:
		observer.upload("", "")
		break
	case DELETE_CLOUD_FILE:
		observer.deleteCloudFile()
		break
	case DELETE_LOCAL_FILE:
		observer.deleteLocalFile()
		break
	case DELETE_CLOUD_FOLDER:
		observer.deleteCloudFolder("")
		break
	case DELETE_LOCAL_FOLDER:
		observer.deleteLocalFolder("")
		break
	case UPDATE_CLOUD_FILE:
		observer.updateCloudFile()
		break
	case UPDATE_LOCAL_FILE:
		observer.updateLocalFile()
		break
	case RENAME_CLOUD_FILE:
		observer.renameCloudFile()
		break
	case RENAME_LOCAL_FILE:
		observer.renameLocalFile()
		break
	case DOWNLOAD_FILE:
		observer.download("", "")
		break
	case RENAME_CLOUD_FOLDER:
		observer.renameCloudFolder()
		break
	case RENAME_LOCAL_FOLDER:
		observer.renameLocalFile()
	default:
		break
	}
}
