package main

import (
	"fmt"
	"github.com/autlunatic/MySqlBackup/ftp"
	"github.com/autlunatic/MySqlBackup/Zipping"
	"github.com/autlunatic/goConfig"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const confFile = "backup.conf"

type ftpConfig struct {
	Host       string
	Username   string
	Password   string `encrypted:"true"`
	LocalPath  string
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
	// Configure a MySQLBackupConf exporter
	var mysql MySQLBackupConf

	file, err := os.OpenFile(confFile, os.O_RDWR, 0666)
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	crw := encryptedConfig.ConfigReadWriter{&mysql, file, "maximumSecurity"}
	crw.DoRead()

	// Export the database, and send it to the bucket in the `db_backups` folder
	fmt.Println("exporting...")
	filename := mysql.Export()
	if filename == "" {
		fmt.Println("filename is empty something went wrong with the dump...")
		return
	}
	fmt.Println("zipping ...")
	files := []string{filename}
	Zipping.ZipFiles(filename+".zip", files)
	fmt.Println("zipping done!")
	// remove the old File because we have the zipped version
	os.Remove(filename)
	// move the zipped file to the specified path
	os.Rename(filename+".zip", mysql.CopyToFilePath+filepath.Base(filename)+".zip")
	// do the upload to ftp servers
	fmt.Println("uploading ... ")
	mysql.uploadFile()
	fmt.Println("All Done!")
}

// Export produces a `mysqldump` of the specified database, and creates a gzip compressed tarball archive.
func (m MySQLBackupConf) Export() string {
	dumpPath := fmt.Sprintf(`bu_%v_%v.sql`, m.DB, time.Now().Unix())

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
func (m MySQLBackupConf) uploadFile() {
	for _, servers := range m.FtpConfig {
		uc := ftp.UploadConf{
			servers.Host,
			servers.Username,
			servers.Password,
			servers.RemotePath,
			servers.LocalPath,
		}
		ftp.UploadFile(uc)
	}
}
