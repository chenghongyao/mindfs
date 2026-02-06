package skills

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
)

type ExecuteResult struct {
	Output string      `json:"output"`
	Data   interface{} `json:"data,omitempty"`
}

func ExecuteSkill(ctx context.Context, skill LoadedSkill, params map[string]any) (ExecuteResult, error) {
	command := skill.Config.Command
	if command == "" {
		return ExecuteResult{}, errors.New("skill handler not configured")
	}
	args := append([]string{}, skill.Config.Args...)
	cmd := exec.CommandContext(ctx, command, args...)
	input, _ := json.Marshal(params)
	cmd.Stdin = bytes.NewReader(input)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return ExecuteResult{}, err
	}
	return ExecuteResult{Output: out.String()}, nil
}
