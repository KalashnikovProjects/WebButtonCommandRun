package filesystem

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/core/data"
	"github.com/gofiber/fiber/v2/log"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Adapter struct{}

func Connect() (Adapter, error) {
	err := os.MkdirAll(config.Config.DataFolderPath, 0750)

	if err != nil {
		return Adapter{}, err
	}
	return Adapter{}, nil
}

func (a Adapter) SaveFile(fileId uint, bytes []byte) error {
	filesDir := filepath.Join(config.Config.DataFolderPath, "files")
	if err := os.MkdirAll(filesDir, 0750); err != nil {
		return err
	}

	filePath := filepath.Join(filesDir, fmt.Sprintf("%d", fileId))

	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func(dst *os.File) {
		err := dst.Close()
		if err != nil {
			log.Warn(err)
		}
	}(dst)

	if _, err := dst.Write(bytes); err != nil {
		return err
	}

	return nil
}

func (a Adapter) ClearFiles() error {
	err := os.RemoveAll(filepath.Join(config.Config.DataFolderPath, "files"))
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(config.Config.DataFolderPath, "files"), 0750)
	if err != nil {
		return err
	}
	return nil
}

func (a Adapter) GetFileData(fileId uint) ([]byte, error) {
	filePath := filepath.Join(config.Config.DataFolderPath, "files", fmt.Sprintf("%d", fileId))
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Warn(err)
		}
	}(file)
	return io.ReadAll(file)
}

func (a Adapter) DeleteFile(fileId uint) error {
	filePath := filepath.Join(config.Config.DataFolderPath, "files", fmt.Sprintf("%d", fileId))
	err := os.Remove(filePath)
	if err != nil {
		return err
	}
	return err
}

func (a Adapter) ImportFilesFromZipArchive(archiveBytes []byte) ([]data.FileData, error) {
	reader, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
	if err != nil {
		return nil, err
	}
	var res []data.FileData
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		path := file.Name
		var commandId int
		var commandName string
		var fileId int
		var fileName string

		pathList := strings.Split(filepath.ToSlash(filepath.Clean(path)), "/")
		if len(pathList) != 3 {
			continue
		}
		_, err = fmt.Sscanf(pathList[1], "Command id %d - %s", &commandId, &commandName)
		if err != nil {
			continue
		}

		parts := strings.SplitN(pathList[2], " - ", -1)
		if len(parts) != 2 {
			continue
		}
		_, err = fmt.Sscanf(parts[0], "Id %d", &fileId)
		if err != nil {
			continue
		}
		fileName = parts[1]
		f, err := file.Open()
		if err != nil {
			return nil, err
		}
		fileBytes, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}
		res = append(res, data.FileData{
			FileId:    uint(fileId),
			CommandId: uint(commandId),
			Bytes:     fileBytes,
			Params: data.FileParams{
				Filename: fileName,
				Size:     file.UncompressedSize64,
			},
		})
	}
	return res, nil
}
