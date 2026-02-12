package context

import (
	"fmt"

	"mindfs/server/internal/fs"
	"mindfs/server/internal/skills"
)

func BuildServerContext(mode string, root fs.RootInfo, currentView *CurrentViewRef) (ServerContext, error) {
	dirCfg, err := skills.LoadDirConfig(root)
	if err != nil {
		// 配置加载失败时使用空配置，不阻断流程
		dirCfg = skills.DirConfig{}
	}
	rootPath, _ := root.RootDir()

	ctx := ServerContext{
		Common: CommonContext{
			RootPath:        rootPath,
			UserDescription: dirCfg.UserDescription,
		},
	}

	switch mode {
	case "view":
		catalog, schema := LoadCatalog()
		apis := LoadAPIList()
		apis = append(apis, LoadWSActions()...)
		examples, err := LoadViewExamples("")
		if err != nil {
			examples = []ViewExample{}
		}
		var viewDef *ViewDefinition
		if currentView != nil {
			viewDef = &ViewDefinition{
				RuleID: currentView.RuleID,
			}
		}
		ctx.View = &ViewContext{
			Catalog:        catalog,
			RegistrySchema: schema,
			ServerAPIs:     apis,
			CurrentView:    viewDef,
			ViewExamples:   SelectExamples(examples, dirCfg.UserDescription, 3),
		}
	case "skill":
		dirSkills, err := LoadDirectorySkills(root)
		if err != nil {
			return ctx, fmt.Errorf("load directory skills: %w", err)
		}
		ctx.Skill = &SkillContext{DirectorySkills: dirSkills}
	}

	return ctx, nil
}
