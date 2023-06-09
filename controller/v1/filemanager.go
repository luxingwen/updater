package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"updater"
)

type FileController struct {
	handler *updater.MessageHandler
}

func NewFileController(handler *updater.MessageHandler) *FileController {
	controller := &FileController{
		handler: handler,
	}
	controller.registerHandlers()
	return controller
}

func (fc *FileController) registerHandlers() {
	fc.handler.RegisterHandler("v1/GetFileInfo", fc.handleGetFileInfo)
	fc.handler.RegisterHandler("v1/DeleteFile", fc.handleDeleteFile)
	fc.handler.RegisterHandler("v1/MoveFile", fc.handleMoveFile)
	fc.handler.RegisterHandler("v1/DownloadFile", fc.handleDownloadFile)
}

func (fc *FileController) handleGetFileInfo(ctx *updater.Context) error {
	//path := string(ctx.Message.Data)
	// fileInfo, err := fc.fileManager.GetFileInfo(path)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (fc *FileController) handleDeleteFile(ctx *updater.Context) error {
	// var req updater.DeleteRequest
	// if err := json.Unmarshal(ctx.), &req); err != nil {
	// 	return err
	// }
	// if err := fc.fileManager.DeleteFile(req.FilePath); err != nil {
	// 	return err
	// }
	return nil
}

func (fc *FileController) handleMoveFile(ctx *updater.Context) error {
	return nil
}

// 执行信息
type ExecuteInfo struct {
	StartTime time.Time `json:"startTime""`
	EndTime   time.Time `json:"endTime"`
	Message   string    `json:"message"`
}

// ... 类似地，实现 handleMoveFile 和 handleDownloadFile ...
func (fc *FileController) handleDownloadFile(ctx *updater.Context) error {

	ctx.Logger.Println("handleDownloadFile taskid:", ctx.Message.TaskId)

	execinfo := ExecuteInfo{
		StartTime: time.Now(),
	}

	var reqmsg updater.DownloadRequest
	if err := json.Unmarshal(ctx.Message.Data, &reqmsg); err != nil {
		execinfo.EndTime = time.Now()
		execinfo.Message = err.Error()
		ctx.JSON(updater.CODE_ERROR, err.Error(), execinfo)
		return err
	}

	if reqmsg.URL == "" && reqmsg.DownLoadPath == "" {
		execinfo.EndTime = time.Now()
		execinfo.Message = "url or downloadpath is empty"
		ctx.JSON(updater.CODE_ERROR, "url or downloadpath is empty", execinfo)
		return fmt.Errorf("url or downloadpath is empty")
	}

	if reqmsg.URL == "" {
		reqmsg.URL = "http://" + ctx.Client.Server.Url.Host + "/api/v1/pkg/" + reqmsg.DownLoadPath
	}

	ctx.Logger.Println("download url:", reqmsg.URL)
	c, cancel := context.WithTimeout(ctx.Ctx, time.Second*time.Duration(reqmsg.Timeout))
	defer cancel()

	// 创建目标文件夹（如果需要）
	ctx.Logger.Println("autoCreateDir:", reqmsg.AutoCreateDir)
	if reqmsg.AutoCreateDir {
		err := os.MkdirAll(filepath.Dir(reqmsg.DestPath), 0755)
		if err != nil {
			execinfo.EndTime = time.Now()
			execinfo.Message = err.Error()
			ctx.JSON(updater.CODE_ERROR, err.Error(), execinfo)
			return err
		}
	}

	ctx.Logger.Println("overwriteExisted:", reqmsg.OverwriteExisted)
	ctx.Logger.Println("destPath:", reqmsg.DestPath)
	// 检查目标文件是否存在
	if _, err := os.Stat(reqmsg.DestPath); err == nil && !reqmsg.OverwriteExisted {
		err = fmt.Errorf("file already exists and overwriteExisted is set to false")
		execinfo.EndTime = time.Now()
		execinfo.Message = err.Error()
		ctx.JSON(updater.CODE_ERROR, err.Error(), execinfo)
		return err
	}

	//ctx.Client.Server.Url

	// 发起 HTTP 请求
	req, err := http.NewRequestWithContext(c, http.MethodGet, reqmsg.URL, nil)
	if err != nil {
		execinfo.EndTime = time.Now()
		execinfo.Message = err.Error()
		ctx.JSON(updater.CODE_ERROR, err.Error(), execinfo)
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// 检查错误类型是否为超时错误
		execinfo.EndTime = time.Now()
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// 处理超时错误
			execinfo.Message = err.Error()
			ctx.JSON(updater.CODE_TIMEOUT, err.Error(), execinfo)
			return fmt.Errorf("request timed out")
		}

		execinfo.Message = err.Error()
		ctx.JSON(updater.CODE_ERROR, err.Error(), execinfo)
		return err
	}
	defer resp.Body.Close()

	// 创建目标文件
	file, err := os.Create(reqmsg.DestPath)
	if err != nil {
		execinfo.EndTime = time.Now()
		execinfo.Message = err.Error()
		ctx.JSON(updater.CODE_ERROR, err.Error(), execinfo)
		return err
	}
	defer file.Close()

	// 将响应的内容写入文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		execinfo.EndTime = time.Now()
		execinfo.Message = err.Error()
		ctx.JSON(updater.CODE_ERROR, err.Error(), execinfo)
		return err
	}

	execinfo.EndTime = time.Now()
	execinfo.Message = "success"
	ctx.JSONSuccess(execinfo)
	return nil
}
