package v1

import (
	"encoding/json"

	"updater"
)

type FileController struct {
	fileManager *updater.FileManager
	handler     *updater.MessageHandler
}

func NewFileController(handler *updater.MessageHandler) *FileController {
	controller := &FileController{
		fileManager: updater.NewFileManager(),
		handler:     handler,
	}
	controller.registerHandlers()
	return controller
}

func (fc *FileController) registerHandlers() {
	fc.handler.RegisterHandler("GetFileInfo", fc.handleGetFileInfo)
	fc.handler.RegisterHandler("DeleteFile", fc.handleDeleteFile)
	fc.handler.RegisterHandler("MoveFile", fc.handleMoveFile)
	fc.handler.RegisterHandler("DownloadFile", fc.handleDownloadFile)
}

func (fc *FileController) handleGetFileInfo(ctx *updater.Context) error {
	path := string(ctx.Message.Data)
	fileInfo, err := fc.fileManager.GetFileInfo(path)
	if err != nil {
		return err
	}
	return

}

func (fc *FileController) handleDeleteFile(ctx *updater.Context) error {
	var req updater.DeleteRequest
	if err := json.Unmarshal(ctx.), &req); err != nil {
		return err
	}
	if err := fc.fileManager.DeleteFile(req.FilePath); err != nil {
		return err
	}
	return
}


func (fc *FileController) handleMoveFile(ctx *updater.Context) error {

}

// ... 类似地，实现 handleMoveFile 和 handleDownloadFile ...
