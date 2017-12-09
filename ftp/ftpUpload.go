package ftp

import (
	"fmt"
	"github.com/dutchcoders/goftp"
	"os"
	"path/filepath"
)

// UploadConf holds the Information needed to do a FTP Upload with UploadFile
type UploadConf struct {
	Host       string
	UserName   string
	Password   string
	RemotePath string
	FileName   string
}

//UploadFile Uploads the file with the given UploadConf via FTP
func UploadFile(c UploadConf) {
	var err error
	var ftp *goftp.FTP

	if ftp, err = goftp.Connect(c.Host); err != nil {
		panic(err)
	}
	defer ftp.Close()
	fmt.Println("Connected To FTP")

	if err = ftp.Login(c.UserName, c.Password); err != nil {
		panic(err)
	}
	fmt.Println("Logged In")

	var curPath string
	if curPath, err = ftp.Pwd(); err != nil {
		panic(err)
	}
	ftp.Cwd(c.RemotePath)
	fmt.Printf("Current path: %s \n", curPath)

	var file *os.File
	if file, err = os.Open(c.FileName); err != nil {
		panic(err)
	}

	if err := ftp.Stor(filepath.Base(c.RemotePath), file); err != nil {
		panic(err)
	}
	fmt.Println(file.Name())
}
