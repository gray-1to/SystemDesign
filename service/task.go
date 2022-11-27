package service

import (
	"net/http"
	"strconv"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
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
		userID := sessions.Default(ctx).Get("user")
 
    // Get tasks in DB
    var tasks []database.Task
		query := "SELECT id, title, comment, created_at, is_done, priority, deadline FROM tasks INNER JOIN ownerships ON task_id = id WHERE user_id = ?"
    switch {
    case kw != "":
				switch{
				case str_is_done != "" && str_is_not_done != "":
					err = db.Select(&tasks, query + " AND title LIKE ?", userID, "%" + kw + "%")
				case str_is_done != "":
					err = db.Select(&tasks, query + " AND title LIKE ? AND is_done = ?", userID, "%" + kw + "%", true)
				case str_is_not_done != "":
					err = db.Select(&tasks, query + " AND title LIKE ? AND is_done = ?", userID, "%" + kw + "%", false)
				default:        
					err = db.Select(&tasks, query + " AND title LIKE ?", userID, "%" + kw + "%")
				}
    default:
			switch{
			case str_is_done != "" && str_is_not_done != "":
        err = db.Select(&tasks, query, userID)
			case str_is_done != "":
        err = db.Select(&tasks, query + " AND is_done = ?", userID, true)
			case str_is_not_done != "":
        err = db.Select(&tasks, query + " AND is_done = ?", userID, false)
			default:
        err = db.Select(&tasks,query, userID)
			}
    }
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }

		// judge danger deadline
		var danger_deadline []bool
		now_time := time.Now()
		for task_index := range tasks{
			var diff = tasks[task_index].Deadline.Sub(now_time)
			danger_deadline = append(danger_deadline, diff.Hours() < 5*24)
		}
 
    // Render tasks
    ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks, "Kw": kw, "IsDone": str_is_done == "checked", "IsNotDone": str_is_not_done == "checked", "DangerDeadline": danger_deadline})
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

	str_year := strconv.Itoa(task.Deadline.Year())
	str_month := strconv.Itoa(int(task.Deadline.Month()))
	str_day := strconv.Itoa(task.Deadline.Day())
	str_deadline := str_year + "-"  + str_month + "-" + str_day

	// Render task
	ctx.HTML(http.StatusOK, "task.html", gin.H{"Task": task, "Deadline": str_deadline})
}

func NewTaskForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"Title": "Task registration"})
}

func RegisterTask(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
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
	// Get task priority
	str_priority, exist := ctx.GetPostForm("priority")
	if !exist {
			Error(http.StatusBadRequest, "No priority is given")(ctx)
			return
	}
	int_priority, error := strconv.Atoi(str_priority)
	if error != nil {
		Error(http.StatusBadRequest, error.Error())(ctx)
		return
	}
	// Get task deadline
	deadline, exist := ctx.GetPostForm("deadline")
	if !exist {
			Error(http.StatusBadRequest, "No deadline is given")(ctx)
			return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}
	tx := db.MustBegin()
	// Create new data with given title on DB
	result, err := tx.Exec("INSERT INTO tasks (title, comment, priority, deadline) VALUES (?, ?, ?, ?)", title, comment, int_priority, deadline)
	if err != nil {
			tx.Rollback()
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}
	taskID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	_, err = tx.Exec("INSERT INTO ownerships (user_id, task_id) VALUES (?, ?)", userID, taskID)
	if err != nil {
			tx.Rollback()
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}
	tx.Commit()
	// Render status
	ctx.Redirect(http.StatusFound, fmt.Sprintf("/task/%d", taskID))
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


	str_year := strconv.Itoa(task.Deadline.Year())
	str_month := strconv.Itoa(int(task.Deadline.Month()))
	str_day := strconv.Itoa(task.Deadline.Day())
	str_deadline := str_year + "-"  + str_month + "-" + str_day

	// Render edit form
	ctx.HTML(http.StatusOK, "form_edit_task.html",
			gin.H{"Title": fmt.Sprintf("Edit task %d", task.ID), "Task": task, "Deadline": str_deadline})
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
	// Get task priority
	str_priority, exist := ctx.GetPostForm("priority")
	if !exist {
			Error(http.StatusBadRequest, "No priority is given")(ctx)
			return
	}
	int_priority, error := strconv.Atoi(str_priority)
	if error != nil {
		Error(http.StatusBadRequest, error.Error())(ctx)
		return
	}
	// Get task comment
	deadline, exist := ctx.GetPostForm("deadline")
	if !exist {
			Error(http.StatusBadRequest, "No deadline is given")(ctx)
			return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}
	// Create new data with given title on DB
	db.Exec("UPDATE tasks SET title = ?, comment = ?, is_done = ?, priority = ?, deadline = ? WHERE id = ?", title, comment, bool_is_done, int_priority, deadline, id)
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
