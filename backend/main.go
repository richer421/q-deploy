package main

import (
	"github.com/richer421/q-deploy/cmd"
)

// 模板项目使用 q-dev 作为模块名，创建新项目时会被替换为用户的模块名

//go:generate go run github.com/swaggo/swag/cmd/swag init -o ./gen/docs --parseDependency
//go:generate go run ./gen/gorm_gen

func main() {
	cmd.Execute()
}
