package service

import (
	"net/http"
	"github.com/gin-gonic/gin"

	database "todolist.go/db"
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
	err = db.Select(&datas, "SELECT id, title, comment, is_done FROM tasks") // Use DB#Get for one entry
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
