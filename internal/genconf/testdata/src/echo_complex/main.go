package main

import (
	"fmt"

	echov4 "github.com/labstack/echo/v4"
)

func main() {
	e4 := echov4.New()

	api := "/api"
	index := "/index"
	e4.POST(api+"/groups", CreateGroup)
	e4.GET("/api/groups", GetGroups)
	e4.GET("/api/groups/:group_id"+"/users", GetGroupUsers)
	e4.GET(index, Index)
	e4.GET("/api/groups/:group_id/users/:user_id/tasks", GetGroupUserTasks)
	e4.GET(fmt.Sprintf("/view/screen%d", 1), View)

	auth := e4.Group("/auth")
	auth.POST("/login", Login)
	auth.POST("/logout", Logout)

	adminAuth := auth.Group("/admin")
	adminAuth.POST("/login", Login)

	e4.Logger.Fatal(e4.Start(":3000"))
}

func CreateGroup(_ echov4.Context) error {
	return nil
}

func GetGroups(_ echov4.Context) error {
	return nil
}

func Index(_ echov4.Context) error {
	return nil
}

func GetGroupUsers(_ echov4.Context) error {
	return nil
}

func GetGroupUserTasks(_ echov4.Context) error {
	return nil
}

func View(_ echov4.Context) error {
	return nil
}

func Login(_ echov4.Context) error {
	return nil
}

func Logout(_ echov4.Context) error {
	return nil
}
