package main

import "fmt"

type CloudFileSystem struct {
	serviceProvider string
	cfsT            cloudFileSystemTencent
	config          map[string]string
}

func (cfs *CloudFileSystem) init(sp string) {
	cfs.serviceProvider = sp
	if cfs.serviceProvider == "tencent" || cfs.serviceProvider == "Tencent" {
		cfs.serviceProvider = "tencent"
		var appId string
		var secretId string
		var secretKey string
		var region string
		var bucketName string
		var historyPath string
		var localPath string
		var cloudPath string
		var bucketURL string
		fmt.Println("Please enter your COS service provider:")
		fmt.Scanln(&sp)
		fmt.Println("Please enter your appid:")
		fmt.Scanln(&appId)
		tencent["app_id"] = appId
		fmt.Println("Please enter your secret id:")
		fmt.Scanln(&secretId)
		tencent["secret_id"] = secretId
		fmt.Println("Please enter your secret key:")
		fmt.Scanln(&secretKey)
		tencent["secret_key"] = secretKey
		fmt.Println("Please enter your region:")
		fmt.Scanln(&region)
		tencent["region"] = region
		fmt.Println("Please enter your bucket name:")
		fmt.Scanln(&bucketName)
		tencent["bucket_name"] = bucketName
		fmt.Println("Please enter your bucket url:")
		fmt.Scanln(&bucketURL)
		tencent["bucketURL"] = bucketURL
		fmt.Println("Please enter your history path: (C:/example/example....)")
		fmt.Scanln(&historyPath)
		tencent["history_path"] = historyPath
		fmt.Println("Please enter your local path: (C:/example/example....)")
		fmt.Scanln(&localPath)
		tencent["local_path"] = localPath
		fmt.Println("Please enter your cloud path: (bucketName/example/example....)")
		fmt.Scanln(&cloudPath)
		tencent["cloud_path"] = cloudPath
		cfs.cfsT.init()
		//initialize the tencent cos-config
		cfs.config = make(map[string]string)
		cfs.config["history_path"] = tencent["history_path"]
		cfs.config["local_path"] = tencent["local_path"]
		cfs.config["cloud_path"] = tencent["cloud_path"]
	}

}

func (cfs *CloudFileSystem) upload(cloudPath string, localPath string) {
	if cfs.serviceProvider == "tencent" {
		cfs.cfsT.upload(cloudPath, localPath)
	}
}

func (cfs *CloudFileSystem) download(cloudPath string, localPath string) {
	if cfs.serviceProvider == "tencent" {
		cfs.cfsT.download(cloudPath, localPath)
	}
}

func (cfs *CloudFileSystem) delete(cloudPath string) {
	if cfs.serviceProvider == "tencent" {
		cfs.cfsT.delete(cloudPath)
	}
}

func (cfs *CloudFileSystem) update(cloudPath string, localPath string) {
	if cfs.serviceProvider == "tencent" {
		cfs.cfsT.update(cloudPath, localPath)
	}
}

func (cfs *CloudFileSystem) rename(oldCloudPath string, newCloudPath string) {
	if cfs.serviceProvider == "tencent" {
		cfs.cfsT.rename(oldCloudPath, newCloudPath)
	}
}

func (cfs *CloudFileSystem) copy(srcPath string, distPath string) {
	if cfs.serviceProvider == "tencent" {
		cfs.cfsT.copy(srcPath, distPath)
	}
}

func (cfs *CloudFileSystem) createFolder(cloudPath string) {
	if cfs.serviceProvider == "tencent" {
		cfs.cfsT.createFolder(cloudPath)
	}
}

func (cfs *CloudFileSystem) listFiles(cloudPath string) []string {
	if cfs.serviceProvider == "tencent" {
		return cfs.cfsT.listFiles(cloudPath)
	} else {
		return nil
	}
}

func (cfs *CloudFileSystem) statFile(cloudPath string) map[string]string {
	if cfs.serviceProvider == "tencent" {
		return cfs.cfsT.statFile(cloudPath)
	} else {
		return nil
	}
}

func (cfs *CloudFileSystem) setStat(cloudPath string, metadata map[string]string) {
	if cfs.serviceProvider == "tencent" {
		cfs.cfsT.setStat(cloudPath, metadata)
	}
}

func (cfs *CloudFileSystem) setHash(cloudPath string, hashValue string) {
	if cfs.serviceProvider == "tencent" {
		cfs.cfsT.setHash(cloudPath, hashValue)
	}
}

func (cfs *CloudFileSystem) setMtime(cloudPath string, mTime string) {
	if cfs.serviceProvider == "tencent" {
		cfs.cfsT.setMtime(cloudPath, mTime)
	}
}
