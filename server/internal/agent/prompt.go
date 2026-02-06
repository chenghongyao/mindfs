package agent

import (
	"encoding/json"
	"strings"

	"mindfs/server/internal/context"
)

func BuildSkillFallbackPrompt(skills []context.SkillBrief) string {
	if len(skills) == 0 {
		return ""
	}
	payload, _ := json.Marshal(skills)
	lines := []string{
		"你可以调用以下目录自定义技能 (通过 POST /api/skills/{id}/execute):",
		string(payload),
	}
	return strings.Join(lines, "\n")
}
