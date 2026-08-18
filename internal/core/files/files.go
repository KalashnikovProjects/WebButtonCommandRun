package files

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	projectErrors "github.com/KalashnikovProjects/WebButtonCommandRun/internal/errors"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/utils"
	"github.com/gofiber/fiber/v2/log"
	"io"
	"os"
	"path/filepath"
)

type Service struct {
	filesDirPath       string
	maxFileSize        int64
	commandsRepository CommandsRepository
	filesRepository    FilesRepository
	filesystem         Filesystem
}

func NewService(filesDirPath string, maxFilesSize int64, commandsRepository CommandsRepository, filesRepository FilesRepository, filesystem Filesystem) *Service {
	return &Service{
		filesDirPath:       filesDirPath,
		maxFileSize:        maxFilesSize,
		commandsRepository: commandsRepository,
		filesRepository:    filesRepository,
		filesystem:         filesystem,
	}
}

func (s Service) validateFile(data *entities.FileParams) error {
	if s.maxFileSize > 0 && int64(data.Size) > s.maxFileSize {
		return projectErrors.ErrFileToBig
	}
	if err := utils.CheckName(data.Filename); err != nil {
		return err
	}
	return nil
}

func (s Service) AppendFile(commandID uint, fileBytes []byte, data *entities.FileParams) error {
	exists, err := s.commandsRepository.CommandExists(commandID)
	if err != nil {
		return err
	}
	if !exists {
		return projectErrors.ErrFileToBig
	}

	if err := s.validateFile(data); err != nil {
		return err
	}
	embeddedFile := entities.EmbeddedFile{
		CommandID: commandID,
		Name:      data.Filename,
	}
	if err := s.filesRepository.AppendFile(&embeddedFile); err != nil {
		return err
	}
	if err := s.filesystem.SaveFile(embeddedFile.ID, fileBytes); err != nil {
		return err
	}
	return nil
}

func (s Service) DeleteFile(commandId, fileId uint) error {
	err := s.filesRepository.DeleteFile(commandId, fileId)
	if err != nil {
		return err
	}
	err = s.filesystem.DeleteFile(fileId)
	if err != nil {
		return err
	}
	return nil
}

func (s Service) PatchFile(commandId, fileId uint, newFile *entities.EmbeddedFile) error {
	if newFile.Name != "" {
		if err := utils.CheckName(newFile.Name); err != nil {
			return err
		}
	}
	err := s.filesRepository.PatchFile(commandId, fileId, newFile)
	if err != nil {
		return err
	}
	return nil
}

func (s Service) PutFile(commandId, fileId uint, newFile *entities.EmbeddedFile) error {
	if err := utils.CheckName(newFile.Name); err != nil {
		return err
	}
	utils.SetDefaultFileName(newFile)
	err := s.filesRepository.UpdateFile(commandId, fileId, newFile)
	if err != nil {
		return err
	}
	return nil
}

func (s Service) GetFile(commandId, fileId uint) (*entities.EmbeddedFile, error) {
	file, err := s.filesRepository.GetFile(commandId, fileId)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (s Service) GetCommandFiles(commandId uint) ([]entities.EmbeddedFile, error) {
	exists, err := s.commandsRepository.CommandExists(commandId)
	if err != nil {
		return nil, fmt.Errorf("cant check command exist: %w", err)
	}
	if !exists {
		return nil, projectErrors.ErrNotFound
	}
	return s.filesRepository.GetCommandFiles(commandId)
}

func (s Service) GetAllFilesList() ([]entities.EmbeddedFile, error) {
	return s.filesRepository.GetAllFiles()
}

func (s Service) DownloadFile(commandId, fileId uint) (*entities.EmbeddedFile, []byte, error) {
	fileData, err := s.GetFile(commandId, fileId)
	if err != nil {
		return nil, nil, err
	}
	data, err := s.filesystem.GetFileData(fileId)
	if err != nil {
		return nil, nil, err
	}
	return fileData, data, err
}

func (s Service) DownloadCommandFilesInArchive(commandId uint) ([]byte, error) {
	filesDatas, err := s.filesRepository.GetCommandFiles(commandId)
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
			sourceFilePath := filepath.Join(s.filesDirPath, fmt.Sprintf("%d", fileData.ID))
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

func (s Service) DownloadAllFilesInArchive() ([]byte, error) {
	filesDatas, err := s.filesRepository.GetAllFilesWithCommandInfo()
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
			sourceFilePath := filepath.Join(s.filesDirPath, fmt.Sprintf("%d", fileData.ID))
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

func (s Service) clearFiles() error {
	err := s.filesRepository.DeleteAllFiles()
	if err != nil {
		return err
	}
	return s.filesystem.ClearFiles()
}

func (s Service) ImportAllFilesFromZipArchive(data []byte) error {
	if err := s.clearFiles(); err != nil {
		return err
	}
	filesToAppend, err := s.filesystem.ImportFilesFromZipArchive(data)
	if err != nil {
		return err
	}
	for _, file := range filesToAppend {
		err = s.AppendFile(file.CommandId, file.Bytes, &entities.FileParams{Filename: file.Params.Filename, Size: file.Params.Size})
		if err != nil {
			return err
		}
	}
	return nil
}
