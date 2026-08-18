package webserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/KalashnikovProjects/WebButtonCommandRun/internal/entities"
	projectErrors "github.com/KalashnikovProjects/WebButtonCommandRun/internal/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"golang.org/x/sync/errgroup"
	"io"
	"mime/multipart"
	"strings"
)

func (s *Server) postFiles() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		commandId, err := c.ParamsInt("command_id")
		if err != nil || commandId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		form, err := c.MultipartForm()
		if err != nil {
			return err
		}
		files := form.File["files"]
		group, ctx := errgroup.WithContext(ctx)
		doBrake := false
		for _, file := range files {
			select {
			case <-ctx.Done():
				doBrake = true
			default:
				group.Go(func() error {
					src, err := file.Open()
					if err != nil {
						return fmt.Errorf("error opening file: %w", err)
					}
					fileBytes, err := io.ReadAll(src)
					if err != nil {
						return err
					}
					defer func(src multipart.File) {
						err := src.Close()
						if err != nil {
							log.Warn(err)
						}
					}(src)
					if err := s.files.AppendFile(uint(commandId), fileBytes, &entities.FileParams{Filename: file.Filename, Size: uint64(file.Size)}); err != nil {
						if errors.Is(err, projectErrors.ErrNotFound) {
							return fiber.ErrNotFound
						}
						if errors.Is(err, projectErrors.ErrFileToBig) {
							return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("too big file (max %d) bytes", s.maxFileSize))
						}
						if errors.Is(err, projectErrors.ErrBadName) {
							return fiber.NewError(fiber.StatusBadRequest, "bad file name")
						}
						log.Error(err)
						return fiber.ErrInternalServerError
					}
					return nil
				})
			}
			if doBrake {
				break
			}
		}
		return group.Wait()
	}
}

func (s *Server) getCommandFilesList() fiber.Handler {
	return func(c *fiber.Ctx) error {
		commandId, err := c.ParamsInt("command_id")
		if err != nil || commandId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		commands, err := s.files.GetCommandFiles(uint(commandId))
		if errors.Is(err, projectErrors.ErrNotFound) {
			return fiber.ErrNotFound
		} else if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(commands)
	}
}

func (s *Server) getFile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		commandId, err := c.ParamsInt("command_id")
		if err != nil || commandId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		fileId, err := c.ParamsInt("file_id")
		if err != nil || fileId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid file id")
		}
		commands, err := s.files.GetFile(uint(commandId), uint(fileId))
		if errors.Is(err, projectErrors.ErrNotFound) {
			return fiber.ErrNotFound
		} else if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(commands)
	}
}

func (s *Server) putFile() fiber.Handler {
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
		err = s.files.PutFile(uint(commandId), uint(fileId), &file)
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

func (s *Server) patchFile() fiber.Handler {
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
		err = s.files.PatchFile(uint(commandId), uint(fileId), &file)
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

func (s *Server) deleteFile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		commandId, err := c.ParamsInt("command_id")
		if err != nil || commandId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		fileId, err := c.ParamsInt("file_id")
		if err != nil || fileId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid file id")
		}
		err = s.files.DeleteFile(uint(commandId), uint(fileId))
		if errors.Is(err, projectErrors.ErrNotFound) {
			return fiber.ErrNotFound
		} else if err != nil {
			return fiber.ErrInternalServerError
		}
		return nil
	}
}

func (s *Server) downloadFile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		commandId, err := c.ParamsInt("command_id")
		if err != nil || commandId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		fileId, err := c.ParamsInt("file_id")
		if err != nil || fileId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid file id")
		}
		fileData, file, err := s.files.DownloadFile(uint(commandId), uint(fileId))
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

func (s *Server) downloadAllFiles() fiber.Handler {
	return func(c *fiber.Ctx) error {
		archive, err := s.files.DownloadAllFilesInArchive()
		if err != nil {
			log.Error(err)
			return fiber.ErrInternalServerError
		}
		c.Type("zip")
		return c.Send(archive)
	}
}

func (s *Server) downloadCommandFiles() fiber.Handler {
	return func(c *fiber.Ctx) error {
		commandId, err := c.ParamsInt("command_id")
		if err != nil || commandId < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid command id")
		}
		archive, err := s.files.DownloadCommandFilesInArchive(uint(commandId))
		if err != nil {
			log.Error(err)
			return fiber.ErrInternalServerError
		}
		c.Type("zip")
		return c.Send(archive)
	}
}

func (s *Server) importFiles() fiber.Handler {
	return func(c *fiber.Ctx) error {
		form, err := c.MultipartForm()
		if err != nil {
			return err
		}
		files := form.File["files"]
		if len(files) == 0 {
			return fiber.ErrBadRequest
		}
		if !strings.HasSuffix(files[0].Filename, ".zip") {
			return fiber.NewError(fiber.StatusBadRequest, "Only zip archives accepted")
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
		err = s.files.ImportAllFilesFromZipArchive(bytes)
		if err != nil {
			return err
		}
		return nil
	}
}
