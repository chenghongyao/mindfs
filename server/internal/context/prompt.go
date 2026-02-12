package context

import (
	"encoding/json"
	"strings"
)

func BuildServerPrompt(mode string, ctx ServerContext) string {
	lines := []string{}
	if ctx.Common.UserDescription != "" {
		lines = append(lines, "当前项目描述: "+ctx.Common.UserDescription)
	}
	if mode == "view" && ctx.View != nil {
		catalog, _ := json.Marshal(ctx.View.Catalog)
		schema, _ := json.Marshal(ctx.View.RegistrySchema)
		apis, _ := json.Marshal(ctx.View.ServerAPIs)
		lines = append(lines, "组件 Catalog: "+string(catalog))
		lines = append(lines, "组件 Schema: "+string(schema))
		lines = append(lines, "可用 API: "+string(apis))
	}
	if mode == "skill" && ctx.Skill != nil {
		payload, _ := json.Marshal(ctx.Skill.DirectorySkills)
		lines = append(lines, "目录技能: "+string(payload))
	}
	if len(lines) == 0 {
		return ""
	}
	return "[ASSIST_CONTEXT]\n" + strings.Join(lines, "\n")
}

func BuildUserPrompt(message string, clientCtx ClientContext) string {
	lines := []string{strings.TrimSpace(message)}
	if clientCtx.CurrentPath != "" {
		lines = append(lines, "选中文件: "+clientCtx.CurrentPath)
	}
	if clientCtx.Selection != nil {
		lines = append(lines, "选中内容: "+clientCtx.Selection.Text)
	}
	return "[USER_INPUT]\n" + strings.Join(lines, "\n")
}
