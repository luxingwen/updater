package updater

import (
	"testing"
)

var scriptContent = `
#!/bin/bash
echo "参数个数：$#"
echo "$1";
pwd;
echo 'Hello, World!';
ls -al
`

func TestScriptTaskRun(t *testing.T) {
	// 创建一个 ScriptTaskRequest 对象
	request := &ScriptTaskRequest{
		TaskID:      "test-task",
		Type:        "test-type",
		Content:     scriptContent,
		Params:      []string{"1234"},
		Timeout:     10, // 设置超时时间（秒）
		Interpreter: "",
		WorkDir:     "/Users/luxingwen/lxw/updater-project/updater",
		Stdin:       "",
	}

	// 创建一个 ScriptTask 对象
	scriptTask := NewScriptTask(request)

	// 调用 Run 方法
	err := scriptTask.Run()
	if err != nil {
		t.Errorf("ScriptTask Run failed: %v", err)
	}

	// 检查结果
	if scriptTask.ScriptResult.Code != CodeSuccess {
		t.Errorf("ScriptTask Run failed with error code: %s", scriptTask.ScriptResult.Code)
	}

	// 打印输出信息
	t.Log("ScriptTask Run stdout:", scriptTask.ScriptResult.Stdout)
	t.Log("ScriptTask Run stderr:", scriptTask.ScriptResult.Stderr)
	t.Log("ScriptTask Run error:", scriptTask.ScriptResult.Error)
	t.Log("ScriptTask Run exit code:", scriptTask.ScriptResult.ExitCode)
}
