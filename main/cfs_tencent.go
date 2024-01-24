package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/tencentyun/cos-go-sdk-v5"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type cloudFileSystemTencent struct {
	secretId    string
	secretKey   string
	region      string
	token       string
	bucketName  string
	appId       string
	bucket      string
	historyPath string
	localPath   string
	cloudPath   string
}

var client *cos.Client

func (cfsT *cloudFileSystemTencent) init() {
	//Initialize the cloud file system
	cfsT.secretId = tencent["secret_id"]
	cfsT.secretKey = tencent["secret_key"]
	cfsT.region = tencent["region"]
	cfsT.bucketName = tencent["bucket_name"]
	cfsT.appId = tencent["app_id"]
	cfsT.bucket = cfsT.bucketName + "-" + cfsT.appId

	//parse the bucket url
	u, _ := url.Parse(tencent["bucketURL"])
	b := &cos.BaseURL{BucketURL: u}
	client = cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfsT.secretId,
			SecretKey: cfsT.secretKey,
		},
	})

	cfsT.historyPath = tencent["history_path"]
	cfsT.localPath = tencent["local_path"]
	cfsT.cloudPath = tencent["cloud_path"]
}

func (cfsT *cloudFileSystemTencent) upload(cloudPath string, localPath string) {
	if localPath[len(localPath)-1] == '/' { //If the localPath is the path of a folder
		cfsT.createFolder(cloudPath)
	} else {
		//First, get the hash value of the file
		fileHash := getLocalFileHash(localPath)
		if fileHash == "" {
			fmt.Println("Cannot open the file")
			return
		}
		fileMtime := strconv.Itoa(int(time.Now().Unix()))
		UUID := uuid.New()
		fileUUID := UUID.String()
		var metaData http.Header
		metaData = make(http.Header)
		metaData.Add("x-cos-meta-hash", fileHash)
		metaData.Add("x-cos-meta-mtime", fileMtime)
		metaData.Add("x-cos-meta-uuid", fileUUID)
		//upload the file
		key := cloudPath
		f, err := os.Open(localPath)
		opt := &cos.ObjectPutOptions{
			ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
				XCosMetaXXX: &metaData,
			},
		}
		_, err = client.Object.Put(context.Background(), key, f, opt)
		/*_, err = client.Object.PutFromFile(context.Background(), key, localPath, opt)*/
		if err != nil {
			panic(err)
		}
	}
}

func (cfsT *cloudFileSystemTencent) download(cloudPath string, localPath string) {
	//Avoid the situation that the local file has already existed.
	tempLocalPath := localPath + (uuid.New().String())
	_, err := client.Object.GetToFile(context.Background(), cloudPath, tempLocalPath, nil)
	if err != nil {
		panic(err)
	}
	_, err = os.Stat(localPath)
	if err == nil || os.IsExist(err) {
		os.Remove(localPath)
	}
	os.Rename(tempLocalPath, localPath)
}

func (cfsT *cloudFileSystemTencent) delete(cloudPath string) {
	_, err := client.Object.Delete(context.Background(), cloudPath)
	if err != nil {
		panic(err)
	}
}

func (cfsT *cloudFileSystemTencent) update(cloudPath string, localPath string) {
	/* Update the file on the cloud.
	You need to set the mtime, hash, uuid
	the mtime is current time
	*/
	ok, err := client.Object.IsExist(context.Background(), cloudPath)
	if cloudPath[len(cloudPath)-1] == '/' || (err == nil && !ok) {
		//The cloudPath is a path of folder or the file doesn't exist on the cloud.
		return
	}
	fileHash := getLocalFileHash(localPath)
	if fileHash == "" {
		panic(fileHash)
	}
	fileMtime := strconv.Itoa(int(time.Now().Unix()))
	fileUUID := cfsT.statFile(cloudPath)["uuid"]
	var metaData http.Header
	metaData = make(http.Header)
	metaData.Add("x-cos-meta-hash", fileHash)
	metaData.Add("x-cos-meta-mtime", fileMtime)
	metaData.Add("x-cos-meta-uuid", fileUUID)
	//upload the file
	key := cloudPath
	f, err := os.Open(localPath)
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			XCosMetaXXX: &metaData,
		},
	}
	_, err = client.Object.Put(context.Background(), key, f, opt)
	/*_, err = client.Object.PutFromFile(context.Background(), key, localPath, opt)*/
	if err != nil {
		panic(err)
	}
}

func (cfsT *cloudFileSystemTencent) rename(oldCloudPath string, newCloudPath string) {
	if oldCloudPath[len(oldCloudPath)-1] == '/' {
		//If the oldCloudPath is a path of folder, the children of it should be renamed relatively.
		for _, filename := range cfsT.listFiles(oldCloudPath) {
			cfsT.rename(oldCloudPath+filename, newCloudPath+filename)
		}
	}
	u, _ := url.Parse(tencent["bucketURL"])
	sourceURL := fmt.Sprintf("%s%s", u.Host, oldCloudPath)
	_, _, err := client.Object.Copy(context.Background(), newCloudPath, sourceURL, nil)
	if err != nil {
		panic(err)
	}
	cfsT.delete(oldCloudPath)
	cfsT.setMtime(newCloudPath, strconv.Itoa(int(time.Now().Unix())))
}

func (cfsT *cloudFileSystemTencent) copy(srcPath string, distPath string) {
	if srcPath[len(srcPath)-1] == '/' {
		//If the oldCloudPath is a path of folder, the children of it should be renamed relatively.
		for _, filename := range cfsT.listFiles(srcPath) {
			cfsT.copy(srcPath+filename, distPath+filename)
		}
	}
	u, _ := url.Parse(tencent["bucketURL"])
	sourceURL := fmt.Sprintf("%s%s", u.Host, srcPath)
	_, _, err := client.Object.Copy(context.Background(), distPath, sourceURL, nil)
	if err != nil {
		panic(err)
	}
	cfsT.setMtime(distPath, strconv.Itoa(int(time.Now().Unix())))
}

func (cfsT *cloudFileSystemTencent) createFolder(cloudPath string) {
	if cloudPath[len(cloudPath)-1] != '/' {
		cloudPath += "/"
	}
	fileHash := getBufferHash([]byte(""))
	fileMtime := strconv.Itoa(int(time.Now().Unix()))
	fileUUID := uuid.New().String()
	var metaData http.Header
	metaData = make(http.Header)
	metaData.Add("x-cos-meta-hash", fileHash)
	metaData.Add("x-cos-meta-mtime", fileMtime)
	metaData.Add("x-cos-meta-uuid", fileUUID)
	//upload the file
	key := cloudPath
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			XCosMetaXXX: &metaData,
		},
	}
	_, err := client.Object.Put(context.Background(), key, strings.NewReader(""), opt)
	/*_, err = client.Object.PutFromFile(context.Background(), key, localPath, opt)*/
	if err != nil {
		panic(err)
	}
}

func (cfsT *cloudFileSystemTencent) listFiles(cloudPath string) []string {
	var files []string
	var marker string
	if cloudPath[len(cloudPath)-1] != '/' {
		cloudPath += "/"
	}
	opt := &cos.BucketGetOptions{
		Prefix:    cloudPath,
		Delimiter: "/",
		MaxKeys:   1000,
	}
	isTruncated := true
	for isTruncated {
		opt.Marker = marker
		v, _, err := client.Bucket.Get(context.Background(), opt)
		if err != nil {
			fmt.Println(err)
			break
		}
		for _, content := range v.Contents {
			if content.Key[len(content.Key)-1] != '/' {
				fileNames := strings.Split(content.Key, "/")
				files = append(files, fileNames[len(fileNames)-1])
			}

		}
		// common prefix 表示表示被 delimiter 截断的路径, 如 delimter 设置为/, common prefix 则表示所有子目录的路径
		for _, commonPrefix := range v.CommonPrefixes {
			files = append(files, commonPrefix[len(cloudPath):])
		}
		isTruncated = v.IsTruncated
		marker = v.NextMarker
	}
	return files
}

func (cfsT *cloudFileSystemTencent) statFile(cloudPath string) map[string]string {
	resp, err := client.Object.Head(context.Background(), cloudPath, nil)
	if err != nil {
		return nil
	}
	setFlag := false
	var metaData map[string]string
	metaData = make(map[string]string)
	temp := resp.Header.Get("x-cos-meta-hash")
	if temp != "" {
		metaData["hash"] = temp
	} else {
		if cloudPath[len(cloudPath)-1] == '/' {
			metaData["hash"] = getBufferHash([]byte(""))
		} else {
			metaData["hash"] = getCloudFileHash(cloudPath)
		}
		setFlag = true
	}
	temp = resp.Header.Get("x-cos-meta-mtime")
	if temp != "" {
		metaData["mtime"] = temp
	} else {
		metaData["mtime"] = strconv.Itoa(int(time.Now().Unix()))
		setFlag = true
	}
	temp = resp.Header.Get("x-cos-meta-uuid")
	if temp != "" {
		metaData["uuid"] = temp
	} else {
		metaData["uuid"] = uuid.New().String()
		setFlag = true
	}
	if setFlag {
		cfsT.setStat(cloudPath, metaData)
	}
	return metaData
}

func (cfsT *cloudFileSystemTencent) setStat(cloudPath string, metadata map[string]string) {
	u, _ := url.Parse(tencent["bucketURL"])
	sourceURL := fmt.Sprintf("%s/%s", u.Host, cloudPath)
	var metaData http.Header
	metaData = make(http.Header)
	metaData.Add("x-cos-meta-hash", metadata["hash"])
	metaData.Add("x-cos-meta-mtime", metadata["mtime"])
	metaData.Add("x-cos-meta-uuid", metadata["uuid"])
	opt := &cos.ObjectCopyOptions{
		&cos.ObjectCopyHeaderOptions{
			XCosMetadataDirective: "Replaced",
			XCosMetaXXX:           &metaData,
		},
		nil,
	}
	_, _, err := client.Object.Copy(context.Background(), cloudPath, sourceURL, opt)
	if err != nil {
		panic(err)
	}
}

func (cfsT *cloudFileSystemTencent) setHash(cloudPath string, hashValue string) {
	stat := cfsT.statFile(cloudPath)
	stat["hash"] = hashValue
	cfsT.setStat(cloudPath, stat)
}

func (cfsT *cloudFileSystemTencent) setMtime(cloudPath string, mTime string) {
	stat := cfsT.statFile(cloudPath)
	stat["mtime"] = mTime
	cfsT.setStat(cloudPath, stat)
}
