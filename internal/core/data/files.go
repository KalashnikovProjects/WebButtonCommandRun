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
)

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

func validateFile(data entities.FileParams) error {
	if config.Config.MaxFileSize > 0 && int64(data.Size) > config.Config.MaxFileSize {
		return projectErrors.ErrFileToBig
	}
	if err := checkName(data.Filename); err != nil {
		return err
	}
	return nil
}

func (s service) AppendFile(commandID uint, fileBytes []byte, data entities.FileParams) error {
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
	if err := s.filesystem.SaveFile(embeddedFile.ID, fileBytes); err != nil {
		return err
	}
	return nil
}

func (s service) DeleteFile(commandId, fileId uint) error {
	err := s.filesRepo.DeleteFile(commandId, fileId)
	if err != nil {
		return err
	}
	err = s.filesystem.DeleteFile(fileId)
	if err != nil {
		return err
	}
	return nil
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
	data, err := s.filesystem.GetFileData(fileId)
	if err != nil {
		return entities.EmbeddedFile{}, nil, err
	}
	return fileData, data, err
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
	return s.filesystem.ClearFiles()
}

func (s service) ImportAllFilesFromZipArchive(data []byte) error {
	if err := s.clearFiles(); err != nil {
		return err
	}
	filesToAppend, err := s.filesystem.ImportFilesFromZipArchive(data)
	if err != nil {
		return err
	}
	for _, file := range filesToAppend {
		err = s.AppendFile(file.CommandId, file.Bytes, entities.FileParams{Filename: file.Params.Filename, Size: file.Params.Size})
		if err != nil {
			return err
		}
	}
	return nil
}
