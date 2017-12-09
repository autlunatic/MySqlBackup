package Zipping

import (
	"archive/zip"
	"io"
	"os"
)

//ZipFiles takes all files of params and zips them to one File called filname
func ZipFiles(filename string, files []string) error {

	newFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newFile.Close()

	zipWriter := zip.NewWriter(newFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {

		zipfile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer zipfile.Close()

		// Get the file information
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Change to deflate to gain better compression
		// see http://golang.org/pkg/archive/zip/#pkg-constants
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, zipfile)
		if err != nil {
			return err
		}
	}
	return nil
}
