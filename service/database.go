package service

import (
	"net/http"
	"github.com/gin-gonic/gin"

	database "todolist.go/db"
	"github.com/gin-contrib/sessions"
)

func ShowUsers(ctx *gin.Context){
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	
	var datas []database.User
	err = db.Select(&datas, "SELECT id, name, password FROM users") // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	ctx.JSON(http.StatusOK, datas)
}

func ShowTasks(ctx *gin.Context){
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	
	var datas []database.Task
	err = db.Select(&datas, "SELECT id, title, comment, is_done, priority FROM tasks") // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	ctx.JSON(http.StatusOK, datas)
}

func ShowOwnerships(ctx *gin.Context){
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	
	var datas []database.Ownership
	err = db.Select(&datas, "SELECT user_id, task_id FROM ownerships") // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	ctx.JSON(http.StatusOK, datas)
}

func ShowCategories(ctx *gin.Context){
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	
	var datas []database.Category
	err = db.Select(&datas, "SELECT id, category_name FROM categories") // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	ctx.JSON(http.StatusOK, datas)
}

func ShowTaskCategories(ctx *gin.Context){
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	
	var datas []database.TaskCategory
	err = db.Select(&datas, "SELECT task_id, category_id FROM task_category") // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	ctx.JSON(http.StatusOK, datas)
}

func ShowAdminLoginPage(ctx *gin.Context){
	ctx.HTML(http.StatusOK, "admin_login.html", gin.H{"Title": "Login"})
}

func AdminLogin(ctx *gin.Context){
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	if(username == "admin") && (password == "password"){
		// セッションの保存
		session := sessions.Default(ctx)
		session.Set("admin_login", true)
		session.Save()

		ctx.HTML(http.StatusOK, "database_list.html", gin.H{"Title": "Login"})
		return
	}else{
		ctx.HTML(http.StatusBadRequest, "admin_login.html", gin.H{"Title": "Login", "Username": username, "Error": "Incorrect password"})
		return
	}
}

func AdminLogout(ctx *gin.Context){
	// セッションの保存
	session := sessions.Default(ctx)
	session.Set("admin_login", false)
	session.Save()
	ctx.Redirect(http.StatusFound, "/")
}

func AdminChech(ctx *gin.Context){
	admin_login := sessions.Default(ctx).Get("admin_login")
	if admin_login != nil && admin_login.(bool){
		ctx.Next()
	}else{
		ctx.Redirect(http.StatusFound, "/admin/login")
		ctx.Abort()
	}
}
