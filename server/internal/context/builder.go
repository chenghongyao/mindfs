package context

import (
	"fmt"

	"mindfs/server/internal/session"
	"mindfs/server/internal/skills"
)

func BuildServerContext(mode, rootPath, managedDir, currentPath string, currentView *CurrentViewRef, store *session.Store) (ServerContext, error) {
	dirCfg, err := skills.LoadDirConfig(managedDir)
	if err != nil {
		// 配置加载失败时使用空配置，不阻断流程
		dirCfg = skills.DirConfig{}
	}

	related, err := FindRelatedSessions(store, currentPath, 3)
	if err != nil {
		// 关联 Session 查找失败时使用空列表
		related = []SessionBrief{}
	}

	ctx := ServerContext{
		Common: CommonContext{
			RootPath:        rootPath,
			UserDescription: dirCfg.UserDescription,
			RelatedSessions: related,
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
				RuleID:  currentView.RuleID,
				Version: currentView.Version,
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
		dirSkills, err := LoadDirectorySkills(managedDir)
		if err != nil {
			return ctx, fmt.Errorf("load directory skills: %w", err)
		}
		ctx.Skill = &SkillContext{DirectorySkills: dirSkills}
	}

	return ctx, nil
}
