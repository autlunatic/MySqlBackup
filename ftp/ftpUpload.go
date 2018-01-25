package ftp

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dutchcoders/goftp"
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
func UploadFile(c UploadConf) (err error) {
	var ftp *goftp.FTP

	if ftp, err = goftp.Connect(c.Host); err != nil {
		return err
	}
	defer ftp.Close()
	//	fmt.Println("Connected To FTP")

	if err = ftp.Login(c.UserName, c.Password); err != nil {
		return err
	}
	//fmt.Println("Logged In")

	//var curPath string
	if _, err := ftp.Pwd(); err != nil {
		return err
	}
	ftp.Cwd(c.RemotePath)
	//fmt.Printf("Current path: %s \n", curPath)

	var file *os.File
	if file, err = os.Open(c.FileName); err != nil {
		return err
	}

	if err := ftp.Stor(filepath.Base(c.FileName), file); err != nil {
		return err
	}
	fmt.Println("finished upload: ", file.Name(), "to", c.Host)
	return nil
}
