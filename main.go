package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"todolist.go/db"
	"todolist.go/service"
)

const port = 8000

func main() {
	// initialize DB connection
	dsn := db.DefaultDSN(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	if err := db.Connect(dsn); err != nil {
		log.Fatal(err)
	}

	// initialize Gin engine
	engine := gin.Default()
	engine.LoadHTMLGlob("views/*.html")

	// prepare session
	store := cookie.NewStore([]byte("my-secret"))
	engine.Use(sessions.Sessions("user-session", store))

	// routing
	engine.Static("/assets", "./assets")
	engine.GET("/", service.Home)
	engine.GET("/list", service.LoginCheck, service.TaskList)
	taskGroup := engine.Group("/task")
	taskGroup.Use(service.LoginCheck)
	{
		engine.GET("/:id", service.ShowTask) // ":id" is a parameter
		// タスクの新規登録
		engine.GET("/new", service.NewTaskForm)
		engine.POST("/new", service.RegisterTask)
		// 既存タスクの編集
		engine.GET("/edit/:id", service.EditTaskForm)
		engine.POST("/edit/:id", service.UpdateTask)
		// 既存タスクの削除
		engine.GET("/delete/:id", service.DeleteTask)
	}


	// ユーザ登録
	engine.GET("/user/new", service.NewUserForm)
	engine.POST("/user/new", service.RegisterUser)

	// ログイン
	engine.GET("/login", service.ShowLoginForm)
	engine.POST("/login", service.Login)
	// ログアウト
	engine.GET("/logout", service.Logout)


	// start server
	engine.Run(fmt.Sprintf(":%d", port))
}
