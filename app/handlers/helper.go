package handlers

import (
	"shopTemplate/app/config"
	"shopTemplate/app/db"
	"shopTemplate/app/models"
	"shopTemplate/app/views/layouts"
	"sort"

	"github.com/a-h/templ"
	"github.com/anthdm/superkit/kit"
)

func RenderWithLayout(kit *kit.Kit, content templ.Component) error {
	var user models.AuthUser
	if authedUser, ok := kit.Auth().(models.AuthUser); ok {
		user = authedUser
	}

	categories := getCategoryTree()

	return kit.Render(layouts.App(user, config.Get(), categories, content))
}

func getCategoryTree() []models.Category {
	var allDbCategories []models.Category
	db.Get().Find(&allDbCategories)

	childrenMap := make(map[uint][]models.Category)
	var rootCategories []models.Category

	for _, c := range allDbCategories {
		if c.ParentID == nil {
			rootCategories = append(rootCategories, c)
		} else {
			childrenMap[*c.ParentID] = append(childrenMap[*c.ParentID], c)
		}
	}

	var buildTree func(cats []models.Category) []models.Category
	buildTree = func(cats []models.Category) []models.Category {
		for i := range cats {
			if subCategories, ok := childrenMap[cats[i].ID]; ok {
				cats[i].SubCategories = buildTree(subCategories)
			}
		}
		return cats
	}

	categories := buildTree(rootCategories)

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Position < categories[j].Position
	})

	for i := range categories {
		sortSubCategories(&categories[i])
	}

	return categories
}

func sortSubCategories(category *models.Category) {
	if len(category.SubCategories) > 0 {
		sort.Slice(category.SubCategories, func(i, j int) bool {
			return category.SubCategories[i].Position < category.SubCategories[j].Position
		})
		for i := range category.SubCategories {
			sortSubCategories(&category.SubCategories[i])
		}
	}
}
