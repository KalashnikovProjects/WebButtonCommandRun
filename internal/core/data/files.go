package data

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	projectErrors "github.com/KalashnikovProjects/WebButtonCommandRun/internal/errors"
	"github.com/gofiber/fiber/v2/log"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

type FileData struct {
	Filename string
	Size     uint64
}

func SetDefaultFilesNames(files []entities.EmbeddedFile) {
	for i := 0; i < len(files); i++ {
		SetDefaultFileName(&files[i])
	}
}

func SetDefaultFileName(file *entities.EmbeddedFile) {
	if file.Name == "" {
		file.Name = RandomFileName()
	}
}

func RandomFileName() string {
	return fmt.Sprintf("File %d", rand.Intn(100))
}

func validateFile(data FileData) error {
	if config.Config.MaxFileSize > 0 && int64(data.Size) > config.Config.MaxFileSize {
		return projectErrors.ErrFileToBig
	}
	if err := checkName(data.Filename); err != nil {
		return err
	}
	return nil
}

func saveFile(fileId uint, file io.Reader) error {
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

	if _, err := dst.ReadFrom(file); err != nil {
		return err
	}

	return nil
}

func (s service) AppendFile(commandID uint, file io.Reader, data FileData) error {
	exists, err := s.commandsRepo.CommandExists(commandID)
	if err != nil {
		return err
	}
	if !exists {
		return projectErrors.ErrFileToBig
	}

	if err := validateFile(data); err != nil {
		return err
	}
	embeddedFile := entities.EmbeddedFile{
		CommandID: commandID,
		Name:      data.Filename,
	}
	if err := s.filesRepo.AppendFile(&embeddedFile); err != nil {
		return err
	}
	if err := saveFile(embeddedFile.ID, file); err != nil {
		return err
	}
	return nil
}

func (s service) DeleteFile(commandId, fileId uint) error {
	err := s.filesRepo.DeleteFile(commandId, fileId)
	if err != nil {
		return err
	}
	err = deleteFile(fileId)
	if err != nil {
		return err
	}
	return nil
}

func deleteFile(fileId uint) error {
	filePath := filepath.Join(config.Config.DataFolderPath, "files", fmt.Sprintf("%d", fileId))
	err := os.Remove(filePath)
	if err != nil {
		return err
	}
	return err
}

func (s service) PatchFile(commandId, fileId uint, newFile entities.EmbeddedFile) error {
	if newFile.Name != "" {
		if err := checkName(newFile.Name); err != nil {
			return err
		}
	}
	err := s.filesRepo.PatchFile(commandId, fileId, &newFile)
	if err != nil {
		return err
	}
	return nil
}

func (s service) PutFile(commandId, fileId uint, newFile entities.EmbeddedFile) error {
	if err := checkName(newFile.Name); err != nil {
		return err
	}
	SetDefaultFileName(&newFile)
	err := s.filesRepo.UpdateFile(commandId, fileId, &newFile)
	if err != nil {
		return err
	}
	return nil
}

func (s service) GetFile(commandId, fileId uint) (entities.EmbeddedFile, error) {
	file, err := s.filesRepo.GetFile(commandId, fileId)
	if err != nil {
		return entities.EmbeddedFile{}, err
	}
	return file, nil
}

func (s service) GetCommandFilesList(commandId uint) ([]entities.EmbeddedFile, error) {
	exists, err := s.commandsRepo.CommandExists(commandId)
	if err != nil {
		return nil, fmt.Errorf("cant check command exist: %w", err)
	}
	if !exists {
		return nil, projectErrors.ErrNotFound
	}
	return s.filesRepo.GetCommandFiles(commandId)
}

func (s service) GetAllFilesList() ([]entities.EmbeddedFile, error) {
	return s.filesRepo.GetAllFiles()
}

func (s service) DownloadFile(commandId, fileId uint) (entities.EmbeddedFile, []byte, error) {
	fileData, err := s.GetFile(commandId, fileId)
	if err != nil {
		return entities.EmbeddedFile{}, nil, err
	}
	filePath := filepath.Join(config.Config.DataFolderPath, "files", fmt.Sprintf("%d", fileId))
	file, err := os.Open(filePath)
	if err != nil {
		return entities.EmbeddedFile{}, nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Warn(err)
		}
	}(file)
	res, err := io.ReadAll(file)
	return fileData, res, err
}

func (s service) DownloadCommandFilesInArchive(commandId uint) ([]byte, error) {
	filesDatas, err := s.filesRepo.GetCommandFiles(commandId)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	zipWriterClosed := false
	defer func(zipWriter *zip.Writer) {
		if !zipWriterClosed {
			err := zipWriter.Close()
			if err != nil {
				log.Warn(err)
			}
			zipWriterClosed = true
		}
	}(zipWriter)
	base := "files"
	for _, fileData := range filesDatas {
		err := func() error {
			fileName := fileData.Name
			resultFileName := fmt.Sprintf("Id %d - %s", fileData.ID, fileName)
			fileInZip, err := zipWriter.Create(filepath.Join(base, resultFileName))
			if err != nil {
				return err
			}
			sourceFilePath := filepath.Join(config.Config.DataFolderPath, "files", fmt.Sprintf("%d", fileData.ID))
			sourceFile, err := os.Open(sourceFilePath)
			if err != nil {
				return err
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					log.Warn("error while closing file", err)
				}
			}(sourceFile)
			_, err = io.Copy(fileInZip, sourceFile)
			if err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	err = zipWriter.Flush()
	if err != nil {
		return nil, err
	}
	if !zipWriterClosed {
		err := zipWriter.Close()
		if err != nil {
			return nil, err
		}
		zipWriterClosed = true
	}
	return buf.Bytes(), nil
}

func (s service) DownloadAllFilesInArchive() ([]byte, error) {
	filesDatas, err := s.filesRepo.GetAllFilesWithCommandInfo()
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	zipWriterClosed := false
	defer func(zipWriter *zip.Writer) {
		if !zipWriterClosed {
			err := zipWriter.Close()
			if err != nil {
				log.Warn(err)
			}
			zipWriterClosed = true
		}
	}(zipWriter)
	base := "files"
	for _, fileData := range filesDatas {
		err := func() error {
			commandName := fileData.Command.Name
			commandDir := fmt.Sprintf("Command id %d - %s", fileData.CommandID, commandName)
			fileName := fileData.Name
			resultFileName := fmt.Sprintf("Id %d - %s", fileData.ID, fileName)
			fileInZip, err := zipWriter.Create(filepath.Join(base, commandDir, resultFileName))
			if err != nil {
				return err
			}
			sourceFilePath := filepath.Join(config.Config.DataFolderPath, "files", fmt.Sprintf("%d", fileData.ID))
			sourceFile, err := os.Open(sourceFilePath)
			if err != nil {
				return err
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					log.Warn("error while closing file", err)
				}
			}(sourceFile)
			_, err = io.Copy(fileInZip, sourceFile)
			if err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	err = zipWriter.Flush()
	if err != nil {
		return nil, err
	}
	if !zipWriterClosed {
		err := zipWriter.Close()
		if err != nil {
			return nil, err
		}
		zipWriterClosed = true
	}
	return buf.Bytes(), nil
}

func (s service) clearFiles() error {
	err := s.filesRepo.DeleteAllFiles()
	if err != nil {
		return err
	}
	err = os.RemoveAll(filepath.Join(config.Config.DataFolderPath, "files"))
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(config.Config.DataFolderPath, "files"), 0750)
	if err != nil {
		return err
	}
	return nil
}

func (s service) ImportAllFilesFromArchive(data []byte) error {
	if err := s.clearFiles(); err != nil {
		return err
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
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
			return err
		}

		err = s.AppendFile(uint(commandId), f, FileData{Filename: fileName, Size: file.UncompressedSize64})
		if err != nil {
			return err
		}
	}
	return nil
}
