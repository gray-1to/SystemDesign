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
	engine.GET("/list/:page_id", service.LoginCheck, service.TaskList)
	engine.GET("/list", service.LoginCheck, service.TaskList)
	// engine.GET("/task/new", service.NewTaskForm)
	taskGroup := engine.Group("/task")
	taskGroup.Use(service.LoginCheck)
	{
		taskGroup.GET("/:id", service.ShowTask) // ":id" is a parameter
		// タスクの新規登録
		taskGroup.GET("/new", service.NewTaskForm)
		taskGroup.POST("/new", service.RegisterTask)
		// 既存タスクの編集
		taskGroup.GET("/edit/:id", service.EditTaskForm)
		taskGroup.POST("/edit/:id", service.UpdateTask)
		// 既存タスクの削除
		taskGroup.GET("/delete/:id", service.DeleteTask)
	}

	engine.POST("/category/new", service.RegisterCategory)


	// ユーザ登録
	engine.GET("/user/new", service.NewUserForm)
	engine.POST("/user/new", service.RegisterUser)

	// ログイン
	engine.GET("/login", service.ShowLoginForm)
	engine.POST("/login", service.Login)
	// ログアウト
	engine.GET("/logout", service.Logout)
	// ユーザー削除
	engine.GET("/delete_user", service.DeleteUser)
	// ユーザー情報変更
	engine.GET("/user/change", service.ShowChangeUserInfoForm)
	engine.POST("/user/change", service.ChangeUserInfo)

	// 管理者
	engine.GET("/admin/login", service.ShowAdminLoginPage)
	engine.POST("/admin/login", service.AdminLogin)
	engine.GET("/admin/logout", service.AdminLogout)

	// データベース
	databaseGroup := engine.Group("/database")
	databaseGroup.Use(service.AdminChech)
	{
		databaseGroup.GET("/users", service.ShowUsers)
		databaseGroup.GET("/tasks", service.ShowTasks)
		databaseGroup.GET("/ownerships", service.ShowOwnerships)
		databaseGroup.GET("/categories", service.ShowCategories)
		databaseGroup.GET("/task_categories", service.ShowTaskCategories)
	}


	// start server
	engine.Run(fmt.Sprintf(":%d", port))
}
