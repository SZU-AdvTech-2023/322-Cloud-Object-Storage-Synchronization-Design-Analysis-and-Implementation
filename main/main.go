package main

func main() {
	var cfs CloudFileSystem
	var sp string

	//fmt.Println("Please enter your COS service provider:")
	//fmt.Scanln(&sp)
	sp = "tencent"
	cfs.init(sp)
	initSynchronization(cfs)

	start()
}
