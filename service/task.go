package service

import (
	"net/http"
	"strconv"
	"fmt"

	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

// TaskList renders list of tasks in DB
func TaskList(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

    // Get query parameter
    kw := ctx.Query("kw")

		str_is_done := ctx.Query("is_done")
		str_is_not_done := ctx.Query("is_not_done")
 
    // Get tasks in DB
    var tasks []database.Task
    switch {
    case kw != "":
				switch{
				case str_is_done != "" && str_is_not_done != "":
					err = db.Select(&tasks, "SELECT * FROM tasks WHERE title LIKE ?", "%" + kw + "%")
				case str_is_done != "":
					err = db.Select(&tasks, "SELECT * FROM tasks WHERE title LIKE ? AND is_done = ?", "%" + kw + "%", true)
				case str_is_not_done != "":
					err = db.Select(&tasks, "SELECT * FROM tasks WHERE title LIKE ? AND is_done = ?", "%" + kw + "%", false)
				default:        
					err = db.Select(&tasks, "SELECT * FROM tasks WHERE title LIKE ?", "%" + kw + "%")
				}
    default:
			switch{
			case str_is_done != "" && str_is_not_done != "":
        err = db.Select(&tasks, "SELECT * FROM tasks")
			case str_is_done != "":
        err = db.Select(&tasks, "SELECT * FROM tasks WHERE is_done = ?", true)
			case str_is_not_done != "":
        err = db.Select(&tasks, "SELECT * FROM tasks WHERE is_done = ?", false)
			default:
        err = db.Select(&tasks, "SELECT * FROM tasks")
			}
    }
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
 
    // Render tasks
    ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks, "Kw": kw, "IsDone": str_is_done == "checked", "IsNotDone": str_is_not_done == "checked"})
}

// ShowTask renders a task with given ID
func ShowTask(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// parse ID given as a parameter
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Get a task with given ID
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Render task
	// ctx.String(http.StatusOK, task.Title)  // Modify it!!
	ctx.HTML(http.StatusOK, "task.html", task)
}

func NewTaskForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"Title": "Task registration"})
}

func RegisterTask(ctx *gin.Context) {
	// Get task title
	title, exist := ctx.GetPostForm("title")
	if !exist {
			Error(http.StatusBadRequest, "No title is given")(ctx)
			return
	}
	// Get task comment
	comment, exist := ctx.GetPostForm("comment")
	if !exist {
			Error(http.StatusBadRequest, "No comment is given")(ctx)
			return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}
	// Create new data with given title on DB
	result, err := db.Exec("INSERT INTO tasks (title, comment) VALUES (?, ?)", title, comment)
	if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}
	// Render status
	path := "/list"  // デフォルトではタスク一覧ページへ戻る
	if id, err := result.LastInsertId(); err == nil {
			path = fmt.Sprintf("/task/%d", id)   // 正常にIDを取得できた場合は /task/<id> へ戻る
	}
	ctx.Redirect(http.StatusFound, path)
}

func EditTaskForm(ctx *gin.Context) {
	// ID の取得
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
			Error(http.StatusBadRequest, err.Error())(ctx)
			return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}
	// Get target task
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id)
	if err != nil {
			Error(http.StatusBadRequest, err.Error())(ctx)
			return
	}
	// Render edit form
	ctx.HTML(http.StatusOK, "form_edit_task.html",
			gin.H{"Title": fmt.Sprintf("Edit task %d", task.ID), "Task": task})
}

func UpdateTask(ctx *gin.Context){
	// ID の取得
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
			Error(http.StatusBadRequest, err.Error())(ctx)
			return
	}
	// Get task title
	title, exist := ctx.GetPostForm("title")
	if !exist {
			Error(http.StatusBadRequest, "No title is given")(ctx)
			return
	}
	// Get task comment
	comment, exist := ctx.GetPostForm("comment")
	if !exist {
			Error(http.StatusBadRequest, "No comment is given")(ctx)
			return
	}
	// Get task is_done
	str_is_done, exist := ctx.GetPostForm("is_done")
	if !exist {
			Error(http.StatusBadRequest, "No is_done is given")(ctx)
			return
	}
	bool_is_done := str_is_done == "t"
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}
	// Create new data with given title on DB
	db.Exec("UPDATE tasks SET title = ?, comment = ?, is_done = ? WHERE id = ?", title, comment, bool_is_done, id)
	if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}
	// Render status
	path := "/list"  // デフォルトではタスク一覧ページへ戻る
	path = fmt.Sprintf("/task/%d", id)
	ctx.Redirect(http.StatusFound, path)
}

func DeleteTask(ctx *gin.Context) {
	// ID の取得
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
			Error(http.StatusBadRequest, err.Error())(ctx)
			return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}
	// Delete the task from DB
	_, err = db.Exec("DELETE FROM tasks WHERE id=?", id)
	if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}
	// Redirect to /list
	ctx.Redirect(http.StatusFound, "/list")
}
