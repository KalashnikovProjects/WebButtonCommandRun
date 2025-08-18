package server

import (
	"errors"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/usecases/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"io"
	"mime/multipart"
	"strings"
)

func (a App) PostFiles(c *fiber.Ctx) error {
	commandId, err := c.ParamsInt("command_id")
	if err != nil || commandId < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
	}
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	files := form.File["files"]
	for _, file := range files {
		err = func() error {
			src, err := file.Open()
			if err != nil {
				return fmt.Errorf("error opening file: %w", err)
			}
			defer func(src multipart.File) {
				err := src.Close()
				if err != nil {
					log.Warn(err)
				}
			}(src)
			if err := a.DB.AppendFile(uint(commandId), src, storage.FileData{Filename: file.Filename, Size: uint64(file.Size)}); err != nil {
				if errors.Is(err, storage.ErrNotFound) {
					return fiber.ErrNotFound
				}
				if errors.Is(err, storage.ErrFileToBig) {
					return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("too big file (max %d) bytes", config.Config.MaxFileSize))
				}
				if errors.Is(err, storage.ErrBadName) {
					return fiber.NewError(fiber.StatusBadRequest, "bad file name")
				}
				log.Error(err)
				return fiber.ErrInternalServerError
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func (a App) GetCommandFilesList(c *fiber.Ctx) error {
	commandId, err := c.ParamsInt("command_id")
	if err != nil || commandId < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
	}
	commands, err := a.DB.GetCommandFilesList(uint(commandId))
	if errors.Is(err, storage.ErrNotFound) {
		return fiber.ErrNotFound
	} else if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(commands)
}

func (a App) GetFile(c *fiber.Ctx) error {
	commandId, err := c.ParamsInt("command_id")
	if err != nil || commandId < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
	}
	fileId, err := c.ParamsInt("file_id")
	if err != nil || fileId < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid file id")
	}
	commands, err := a.DB.GetFile(uint(commandId), uint(fileId))
	if errors.Is(err, storage.ErrNotFound) {
		return fiber.ErrNotFound
	} else if err != nil {
		return fiber.ErrInternalServerError
	}
	return c.JSON(commands)
}

func (a App) PutFile(c *fiber.Ctx) error {
	commandId, err := c.ParamsInt("command_id")
	if err != nil || commandId < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
	}
	fileId, err := c.ParamsInt("file_id")
	if err != nil || fileId < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid file id")
	}
	var file entities.EmbeddedFile
	err = c.BodyParser(&file)

	err = a.DB.PutFile(uint(commandId), uint(fileId), file)
	if errors.Is(err, storage.ErrNotFound) {
		return fiber.ErrNotFound
	} else if errors.Is(err, storage.ErrBadName) {
		return fiber.NewError(fiber.StatusBadRequest, "bad file name")
	} else if err != nil {
		return fiber.ErrInternalServerError
	}
	return nil

}

func (a App) PatchFile(c *fiber.Ctx) error {
	commandId, err := c.ParamsInt("command_id")
	if err != nil || commandId < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
	}
	fileId, err := c.ParamsInt("file_id")
	if err != nil || fileId < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid file id")
	}
	var file entities.EmbeddedFile
	err = c.BodyParser(&file)

	err = a.DB.PatchFile(uint(commandId), uint(fileId), file)
	if errors.Is(err, storage.ErrNotFound) {
		return fiber.ErrNotFound
	} else if errors.Is(err, storage.ErrBadName) {
		return fiber.NewError(fiber.StatusBadRequest, "bad file name")
	} else if err != nil {
		return fiber.ErrInternalServerError
	}
	return nil
}

func (a App) DeleteFile(c *fiber.Ctx) error {
	commandId, err := c.ParamsInt("command_id")
	if err != nil || commandId < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
	}
	fileId, err := c.ParamsInt("file_id")
	if err != nil || fileId < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid file id")
	}
	err = a.DB.DeleteFile(uint(commandId), uint(fileId))
	if errors.Is(err, storage.ErrNotFound) {
		return fiber.ErrNotFound
	} else if err != nil {
		return fiber.ErrInternalServerError
	}
	return nil
}

func (a App) DownloadFile(c *fiber.Ctx) error {
	commandId, err := c.ParamsInt("command_id")
	if err != nil || commandId < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
	}
	fileId, err := c.ParamsInt("file_id")
	if err != nil || fileId < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid file id")
	}
	fileData, file, err := a.DB.DownloadFile(uint(commandId), uint(fileId))
	if err != nil {
		return fiber.ErrInternalServerError
	}
	extension := strings.Split(fileData.Name, ".")[0]
	c.Type(extension)

	err = c.Send(file)
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return nil
}

func (a App) DownloadAllFiles(c *fiber.Ctx) error {
	archive, err := a.DB.DownloadAllFilesInArchive()
	if err != nil {
		log.Error(err)
		return fiber.ErrInternalServerError
	}
	c.Type("zip")
	return c.Send(archive)
}

func (a App) DownloadCommandFiles(c *fiber.Ctx) error {
	commandId, err := c.ParamsInt("command_id")
	if err != nil || commandId < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
	}
	archive, err := a.DB.DownloadCommandFilesInArchive(uint(commandId))
	if err != nil {
		log.Error(err)
		return fiber.ErrInternalServerError
	}
	c.Type("zip")
	return c.Send(archive)
}

func (a App) ImportFiles(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	files := form.File["files"]
	if len(files) == 0 {
		return fiber.ErrBadRequest
	}
	src, err := files[0].Open()
	if err != nil {
		return fiber.ErrInternalServerError
	}
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {
			log.Warn(err)
		}
	}(src)
	bytes, err := io.ReadAll(src)
	if err != nil {
		return fiber.ErrInternalServerError
	}
	err = a.DB.ImportAllFilesFromArchive(bytes)
	return nil
}
