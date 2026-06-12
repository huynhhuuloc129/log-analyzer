package utility

import (
	"fmt"
	"io"
	"os"
)

func SaveFile(file interface{ Read([]byte) (int, error) }, filename string, foldername string) error {
	// Create the file on disk

	if err := os.MkdirAll("./"+foldername, os.ModePerm); err != nil {
		return err
	}

	dst, err := os.Create(fmt.Sprintf("./%s/%s", foldername, filename))
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy contents into it
	_, err = io.Copy(dst, file)
	return err
}
