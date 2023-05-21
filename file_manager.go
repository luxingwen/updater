package updater

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type FileInfo struct {
	Name    string `json:"name"`    // 文件名
	Size    int64  `json:"size"`    // 文件大小
	Mode    string `json:"mode"`    // 文件权限
	ModTime int64  `json:"modTime"` // 修改时间
	IsDir   bool   `json:"isDir"`   // 是否是目录
}

func NewFileInfo(fileInfo os.FileInfo) *FileInfo {
	return &FileInfo{
		Name:    fileInfo.Name(),
		Size:    fileInfo.Size(),
		Mode:    fileInfo.Mode().String(),
		ModTime: fileInfo.ModTime().Unix(),
		IsDir:   fileInfo.IsDir(),
	}
}

type FileManager struct{}

func NewFileManager() *FileManager {
	return &FileManager{}
}

// GetFileInfo 获取一个文件或者目录的信息
func (fm *FileManager) GetFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// DeleteFile 删除一个文件
func (fm *FileManager) DeleteFile(path string) error {
	return os.Remove(path)
}

// MoveFile 移动一个文件
func (fm *FileManager) MoveFile(src, dest string) error {
	return os.Rename(src, dest)
}

// BackupFile 备份一个文件
func (fm *FileManager) BackupFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

// DownloadRequest 是下载请求参数
type DownloadRequest struct {
	URL              string `json:"url"`              // 下载 URL
	DestPath         string `json:"destPath"`         // 目标路径
	AutoCreateDir    bool   `json:"autoCreateDir"`    // 是否自动创建文件夹
	OverwriteExisted bool   `json:"overwriteExisted"` // 文件存在是否覆盖文件
}

// DownloadFile 从URL下载文件
func (fm *FileManager) DownloadFile(req DownloadRequest) error {
	// 判断目标文件是否存在
	_, err := os.Stat(req.DestPath)
	if !os.IsNotExist(err) && !req.OverwriteExisted {
		// 如果文件已经存在并且不覆盖，返回错误
		return os.ErrExist
	}

	// 判断是否自动创建文件夹
	if req.AutoCreateDir {
		dir := filepath.Dir(req.DestPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, os.ModePerm)
		}
	}

	// Get the data
	resp, err := http.Get(req.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(req.DestPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
