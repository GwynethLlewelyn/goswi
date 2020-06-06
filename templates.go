package main

import (
    "github.com/gin-contrib/multitemplate"
)

func createMyRender() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("index", "./templates/index.tpl", "./templates/header.tpl", "./templates/navigation.tpl", "./templates/topbar.tpl", "./templates/footer.tpl")
	r.AddFromFiles("welcome", "./templates/welcome.tpl", "./templates/header.tpl", "./templates/navigation.tpl", "./templates/topbar.tpl", "./templates/footer.tpl")
	r.AddFromFiles("404", "./templates/404.tpl", "./templates/header.tpl", "./templates/navigation.tpl", "./templates/topbar.tpl", "./templates/footer.tpl")
	return r
}