package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/autlunatic/goConfig"
	"github.com/autlunatic/goUtil/Zipping"
	"github.com/autlunatic/goUtil/ftp"
)

const confFile = "backup.conf"

type ftpConfig struct {
	Host       string
	Username   string
	Password   string `encrypted:"true"`
	RemotePath string
}

// MySQLBackupConf is an `Exporter` interface that backs up a MySQLBackupConf database via the `mysqldump` command
type MySQLBackupConf struct {
	// DB Host (e.g. 127.0.0.1)
	Host string
	// DB Port (e.g. 3306)
	Port string
	// DB Name
	DB string
	// DB User
	User string
	// DB Password
	Password string `encrypted:"true"`
	// Extra mysqldump options
	// e.g []string{"--extended-insert"}
	Options []string `json:"-"`
	// the path where the Filename is stored
	CopyToFilePath string
	// 	Backup-File should be uploaded to following Ftp servers
	FtpConfig []ftpConfig
	// full path to the mysqldump executable
	MySqlDumpPath string
}

func main() {
	exePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println("Own FilePath not found!", err)
		return
	}
	// Configure a MySQLBackupConf exporter
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error in the script, check your Config! ERROR: ", r)
		}
	}()
	var mysql MySQLBackupConf

	file, err := os.OpenFile(path.Join(exePath, confFile), os.O_RDWR, 0666)
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ConfFile opened", file.Name())

	crw := encryptedConfig.ConfigReadWriter{&mysql, file, "maximumSecurity"}
	err = crw.DoRead()
	if err != nil {
		fmt.Println("Config Read error: ", err)
	}

	if _, err := os.Stat(mysql.CopyToFilePath); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(mysql.CopyToFilePath, os.ModePerm)
			fmt.Println("mkdir", mysql.CopyToFilePath)
		} else {
			fmt.Println("PathStatus Error", err)
		}
	}
	fmt.Println("path found: ", mysql.CopyToFilePath)
	// Export the database, and send it to the bucket in the `db_backups` folder
	fmt.Println("exporting...")
	filename := mysql.Export()
	if filename == "" {
		fmt.Println("filename is empty something went wrong with the dump...")
		return
	}
	fmt.Println("zipping to ", filename+".zip", "...")
	files := []string{filename}
	Zipping.ZipFiles(filename+".zip", files)
	fmt.Println("zipping done!")
	// remove the old File because we have the zipped version
	err = os.Remove(filename)
	if err != nil {
		fmt.Println("error removing file: ", err)
		return
	}
	fmt.Println("Old File deleted!")
	// do the upload to ftp servers
	fmt.Println("uploading ... ")
	err = mysql.uploadFile(mysql.CopyToFilePath + filepath.Base(filename) + ".zip")
	if err != nil {
		fmt.Println("error Uploading file: ", err)
		return
	}
	fmt.Println("All Done!")
}

// Export produces a `mysqldump` of the specified database, and creates a gzip compressed tarball archive.
func (m MySQLBackupConf) Export() string {
	dumpPath := fmt.Sprintf(`bu_%v_%v.sql`, m.DB, time.Now().Format("20060102_150405"))
	dumpPath = path.Join(m.CopyToFilePath, dumpPath)

	options := append(m.dumpOptions(), fmt.Sprintf(`-r%v`, dumpPath))
	_, err := exec.Command(m.MySqlDumpPath, options...).Output()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return dumpPath
}

func (m MySQLBackupConf) dumpOptions() []string {
	options := m.Options
	options = append(options, fmt.Sprintf(`-h%v`, m.Host))
	options = append(options, fmt.Sprintf(`-P%v`, m.Port))
	options = append(options, fmt.Sprintf(`-u%v`, m.User))
	if m.Password != "" {
		options = append(options, fmt.Sprintf(`-p%v`, m.Password))
	}
	options = append(options, m.DB)
	return options
}
func (m MySQLBackupConf) uploadFile(fileName string) error {
	for _, servers := range m.FtpConfig {
		uc := ftp.UploadConf{
			Host:       servers.Host,
			UserName:   servers.Username,
			Password:   servers.Password,
			RemotePath: servers.RemotePath,
			FileName:   fileName,
		}
		err := ftp.UploadFile(uc)
		if err != nil {
			return err
		}
	}
	return nil
}
