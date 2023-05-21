package updater

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ScriptTask struct {
	TaskID          string
	Type            string
	Content         string
	InterpreterArgs []string
	Params          []string
	Timeout         time.Duration
	Interpreter     string
	Stdin           string
	Created         time.Time
	Updated         time.Time
	Status          TaskStatus
	Suffix          string
	Cancel          context.CancelFunc
	ScriptResult    *ScriptResult
	Env             map[string]string
	MachineID       string
}

type ScriptErrorCode string

const (
	CodeCreateTempFileFailed ScriptErrorCode = "CREATE_TEMP_FILE_FAILED"
	CodeWriteTempFileFailed  ScriptErrorCode = "WRITE_TEMP_FILE_FAILED"
	CodeCloseTempFileFailed  ScriptErrorCode = "CLOSE_TEMP_FILE_FAILED"
	CodeChmodTempFileFailed  ScriptErrorCode = "CHMOD_TEMP_FILE_FAILED"
	CodeStartFailed          ScriptErrorCode = "START_FAILED"
	CodeTimeout              ScriptErrorCode = "TIMEOUT"
	CodeStopped              ScriptErrorCode = "STOPPED"
	CodeSuccess              ScriptErrorCode = "SUCCESS"
)

type ScriptResult struct {
	Code      ScriptErrorCode
	Stdout    string
	Stderr    string
	Error     string
	ExitCode  int
	StartTime time.Time
	EndTime   time.Time
}

type ScriptTaskRequest struct {
	TaskID      string            `json:"task_id"`
	Type        string            `json:"type"`
	Content     string            `json:"content"`
	Params      map[string]string `json:"params"`
	Timeout     int               `json:"timeout"`
	Interpreter string            `json:"interpreter"`
	Stdin       string            `json:"stdin"`
}

func NewScriptTask(request *ScriptTaskRequest) *ScriptTask {
	return &ScriptTask{
		TaskID:       request.TaskID,
		Type:         request.Type,
		Content:      request.Content,
		Interpreter:  request.Interpreter,
		Stdin:        request.Stdin,
		Status:       TaskStatusCreated,
		Created:      time.Now(),
		Updated:      time.Now(),
		ScriptResult: &ScriptResult{},
	}
}

func (st *ScriptTask) GetType() string {
	return st.Type
}

func (st *ScriptTask) GetStatus() TaskStatus {
	return st.Status
}

func (st *ScriptTask) GetContent() []byte {
	return []byte(st.Content)
}

func (st *ScriptTask) Run() (err error) {
	defer func() {
		if st.ScriptResult.Code != CodeSuccess {
			st.Status = TaskStatusFailed
		}
	}()

	if len(st.Interpreter) == 0 {
		st.Interpreter = defaultInterpreter
	}
	if len(st.InterpreterArgs) == 0 {
		st.InterpreterArgs = []string{defaultInterpreterArg}
	}
	if st.Suffix == "" {
		st.Suffix = defaultScriptSuffix
	}

	tmpfile, err := ioutil.TempFile("", st.TaskID+st.Suffix)
	if err != nil {
		st.ScriptResult.Error = err.Error()
		st.ScriptResult.Code = CodeCreateTempFileFailed
		return
	}
	defer os.Remove(tmpfile.Name())

	if _, err = tmpfile.Write([]byte(st.Content)); err != nil {
		st.ScriptResult.Error = err.Error()
		st.ScriptResult.Code = CodeWriteTempFileFailed
		return
	}
	if err = tmpfile.Close(); err != nil {
		st.ScriptResult.Error = err.Error()
		st.ScriptResult.Code = CodeCloseTempFileFailed
		return
	}

	err = os.Chmod(tmpfile.Name(), 0755)
	if err != nil {
		st.ScriptResult.Error = err.Error()
		st.ScriptResult.Code = CodeChmodTempFileFailed
		return
	}

	args := append(st.InterpreterArgs, tmpfile.Name())
	args = append(args, st.Params...)

	ctx, cancel := context.WithTimeout(context.Background(), st.Timeout)
	defer cancel()
	st.Cancel = cancel

	cmd := exec.CommandContext(ctx, st.Interpreter, args...)

	if len(st.Stdin) > 0 {
		cmd.Stdin = bytes.NewBufferString(st.Stdin)
	}

	env := make([]string, 0, len(st.Env))
	for k, v := range st.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = append(os.Environ(), env...)

	stdoutPipeReader, stdoutPipeWriter := io.Pipe()
	stderrPipeReader, stderrPipeWriter := io.Pipe()

	cmd.Stdout = stdoutPipeWriter
	cmd.Stderr = stderrPipeWriter

	var stdout, stderr strings.Builder
	stdoutDone := make(chan struct{})
	stderrDone := make(chan struct{})

	go func() {
		defer close(stdoutDone)
		defer stdoutPipeReader.Close()

		reader := bufio.NewReader(stdoutPipeReader)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			st.ScriptResult.Stdout += line
		}
	}()

	go func() {
		defer close(stderrDone)
		defer stderrPipeReader.Close()

		reader := bufio.NewReader(stderrPipeReader)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			st.ScriptResult.Stderr += line
		}
	}()
	startTime := time.Now()
	err = cmd.Start()
	if err != nil {
		st.ScriptResult.Error = err.Error()
		st.ScriptResult.Code = CodeStartFailed
		return
	}

	err = cmd.Wait()
	endTime := time.Now()

	stdoutPipeWriter.Close()
	stderrPipeWriter.Close()

	<-stdoutDone
	<-stderrDone

	var exitCode int
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	if errors.Is(err, context.DeadlineExceeded) {
		st.ScriptResult.Error = "script execution timeout"
		st.ScriptResult.Code = CodeTimeout

		return
	}

	// 添加错误信息到 ScriptResult
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
		fmt.Println("err:", errorMsg)
	}

	fmt.Println(stdout.String())
	fmt.Println(stderr.String())

	st.ScriptResult.Code = CodeSuccess
	st.ScriptResult.EndTime = endTime
	st.ScriptResult.StartTime = startTime
	st.ScriptResult.ExitCode = exitCode
	st.ScriptResult.Error = errorMsg

	return nil
}

func (st *ScriptTask) Stop() error {
	// 实现停止脚本的逻辑
	st.Cancel()
	return nil
}

func (st *ScriptTask) SetStatus(status TaskStatus) {
	st.Status = status
	st.Updated = time.Now()
}
