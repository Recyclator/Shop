// Renders real templates for isolated UI checks without a database or credentials.
package main

import (
	"context"
	"github.com/a-h/templ"
	"github.com/nexora/backend/views/layout"
	"github.com/nexora/backend/views/pages"
	"os"
	"path/filepath"
)

func main() {
	components := map[string]templ.Component{
		"auth":              pages.AuthPage(),
		"customers":         pages.CustomersPage(pages.CustomersData{InitialPage: 1, InitialLimit: 20}),
		"categories":        pages.CategoriesPage(),
		"users":             pages.UsersPage(pages.UsersData{InitialPage: 1}),
		"products":          pages.ProductsPage(),
		"layaways":          pages.LayawaysPage(),
		"pos":               pages.POSPage(),
		"profile":           pages.ProfilePage(),
		"settings":          pages.SettingsPage(),
		"roles-permissions": pages.RolesPermissionsPage(),
		"dashboard":         pages.DashboardPage(pages.DashboardData{}),
	}
	for name, component := range components {
		if name != "auth" {
			component = layout.Base(name, "/"+name, component)
		}
		f, err := os.Create(filepath.Join(os.Args[1], name+".html"))
		if err != nil {
			panic(err)
		}
		if err = component.Render(context.Background(), f); err != nil {
			panic(err)
		}
		f.Close()
	}
}
