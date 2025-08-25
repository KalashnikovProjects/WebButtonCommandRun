package webserver

import (
	"errors"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/config"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/core/data"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	projectErrors "github.com/KalashnikovProjects/WebButtonCommandRun/internal/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"io"
	"mime/multipart"
	"strings"
)

func PostFiles(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
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
				if err := s.Data.AppendFile(uint(commandId), src, data.FileData{Filename: file.Filename, Size: uint64(file.Size)}); err != nil {
					if errors.Is(err, projectErrors.ErrNotFound) {
						return fiber.ErrNotFound
					}
					if errors.Is(err, projectErrors.ErrFileToBig) {
						return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("too big file (max %d) bytes", config.Config.MaxFileSize))
					}
					if errors.Is(err, projectErrors.ErrBadName) {
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
}

func GetCommandFilesList(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
		commandId, err := c.ParamsInt("command_id")
		if err != nil || commandId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		commands, err := s.Data.GetCommandFilesList(uint(commandId))
		if errors.Is(err, projectErrors.ErrNotFound) {
			return fiber.ErrNotFound
		} else if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(commands)
	}
}

func GetFile(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
		commandId, err := c.ParamsInt("command_id")
		if err != nil || commandId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		fileId, err := c.ParamsInt("file_id")
		if err != nil || fileId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid file id")
		}
		commands, err := s.Data.GetFile(uint(commandId), uint(fileId))
		if errors.Is(err, projectErrors.ErrNotFound) {
			return fiber.ErrNotFound
		} else if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(commands)
	}
}

func PutFile(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
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
		if err != nil {
			return err
		}
		err = s.Data.PutFile(uint(commandId), uint(fileId), file)
		if errors.Is(err, projectErrors.ErrNotFound) {
			return fiber.ErrNotFound
		} else if errors.Is(err, projectErrors.ErrBadName) {
			return fiber.NewError(fiber.StatusBadRequest, "bad file name")
		} else if err != nil {
			return fiber.ErrInternalServerError
		}
		return nil
	}
}

func PatchFile(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
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
		if err != nil {
			return err
		}
		err = s.Data.PatchFile(uint(commandId), uint(fileId), file)
		if errors.Is(err, projectErrors.ErrNotFound) {
			return fiber.ErrNotFound
		} else if errors.Is(err, projectErrors.ErrBadName) {
			return fiber.NewError(fiber.StatusBadRequest, "bad file name")
		} else if err != nil {
			return fiber.ErrInternalServerError
		}
		return nil
	}
}

func DeleteFile(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
		commandId, err := c.ParamsInt("command_id")
		if err != nil || commandId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		fileId, err := c.ParamsInt("file_id")
		if err != nil || fileId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid file id")
		}
		err = s.Data.DeleteFile(uint(commandId), uint(fileId))
		if errors.Is(err, projectErrors.ErrNotFound) {
			return fiber.ErrNotFound
		} else if err != nil {
			return fiber.ErrInternalServerError
		}
		return nil
	}
}

func DownloadFile(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
		commandId, err := c.ParamsInt("command_id")
		if err != nil || commandId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		fileId, err := c.ParamsInt("file_id")
		if err != nil || fileId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid file id")
		}
		fileData, file, err := s.Data.DownloadFile(uint(commandId), uint(fileId))
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
}

func DownloadAllFiles(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
		archive, err := s.Data.DownloadAllFilesInArchive()
		if err != nil {
			log.Error(err)
			return fiber.ErrInternalServerError
		}
		c.Type("zip")
		return c.Send(archive)
	}
}

func DownloadCommandFiles(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
		commandId, err := c.ParamsInt("command_id")
		if err != nil || commandId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		archive, err := s.Data.DownloadCommandFilesInArchive(uint(commandId))
		if err != nil {
			log.Error(err)
			return fiber.ErrInternalServerError
		}
		c.Type("zip")
		return c.Send(archive)
	}
}

func ImportFiles(s Services) fiber.Handler {
	return func(c *fiber.Ctx) error {
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
		err = s.Data.ImportAllFilesFromArchive(bytes)
		if err != nil {
			return err
		}
		return nil
	}
}
