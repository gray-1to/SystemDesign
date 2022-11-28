package service

import (
	"net/http"
	"github.com/gin-gonic/gin"

	database "todolist.go/db"
)

func RegisterCategory(ctx *gin.Context){
	// Get category_name
	category_name, exist := ctx.GetPostForm("category_name")
	if !exist {
			Error(http.StatusBadRequest, "No category_name is given")(ctx)
			return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Create new data with given category_name on DB
	_, err = db.Exec("INSERT INTO categories (category_name) VALUES (?)", category_name)
	if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}

	TaskList(ctx)
}
