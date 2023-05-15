package main

import (
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.POST("/api/users", CreateUser)
	e.GET("/api/users", GetUsers)
	e.GET("/api/users/:id", GetUser)
	e.GET("/api/items", GetItems)

	e.Logger.Fatal(e.Start(":3000"))
}

func CreateUser(_ echo.Context) error {
	return nil
}

func GetUsers(_ echo.Context) error {
	return nil
}

func GetUser(_ echo.Context) error {
	return nil
}

func GetItems(_ echo.Context) error {
	return nil
}
