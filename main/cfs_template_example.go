package main

/*

	Cloud Operation Methods Summary


*/

type cloudFileSystemOperations interface {
	init()
	/*
		Init the file system of cloud
		It needs to init the client of COS and certain bucket
	*/
	upload(cloudPath string, localPath string)
	/*
		upload the file
	*/
	download(cloudPath string, localPath string)
	/*
		download the file
	*/
	delete(cloudPath string)
	/*
		delete the file
	*/
	update(cloudPath string, localPath string)
	/*
		update the file
	*/
	rename(oldCloudPath string, newCloudPath string)
	/*
		rename the file
		If the path is a path of directory,the function will be recursively called in turn
		When the rename operation being taken, first, the old file will be copied with new name and the deleted.
	*/
	copy(srcPath string, distPath string)
	/*
		copy the file
	*/
	createFolder(cloudPath string)
	/*
		create a new folder
	*/
	listFiles(cloudPath string) []string
	/*
		List the Subdirectories and files
		Return the array consists of the names of directory/subfile
	*/
	statFile(cloudPath string) map[string]string
	/*
		Get the meta message of the file
		If the file doesn't exist,return nil
		If the meta message involves nil, set it
		Return the meta message(with a map), noticing the type of the mtime is int,here is string, you have to convert it.
	*/
	setStat(cloudPath string, metaDta map[string]string)
	setHash(cloudPath string, hashValue string)
	setMtime(cloudPath string, mTime int)
}
