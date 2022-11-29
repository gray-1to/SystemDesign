package service

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
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
		space_linked_category_names := ctx.Query("category_name")
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

		var category_proper_tasks []database.Task
		// judge category
		if space_linked_category_names != "" {
			var category_ids []int
			category_names := strings.Split(space_linked_category_names, " ")
			// 各カテゴリに対してタスクがカテゴリを持つかチェック
			for _, category_name := range category_names{
				var category_id int
				err = db.Get(&category_id, "SELECT id FROM categories WHERE category_name = ?",category_name)
				if err != nil {
						Error(http.StatusInternalServerError, err.Error())(ctx)
						return
				}
				category_ids = append(category_ids, category_id)
			}
			// 各タスクをチェック
			for _, task := range tasks{
				var hit_num_sum int
				for _, category_id := range category_ids{
					var hit_num int
					err = db.Get(&hit_num, "SELECT COUNT(*) FROM task_category WHERE task_id = ? AND category_id = ?", task.ID, category_id)
					hit_num_sum += hit_num
				}
				if hit_num_sum > 0{
					// 該当タスクが該当カテゴリを持っていた場合
					category_proper_tasks = append(category_proper_tasks, task)
				}
			} 
		}else{
			category_proper_tasks = tasks
		}

		// judge danger deadline
		// clac rest day
		var danger_deadline []bool
		var rest_day []int
		now_time := time.Now()
		for task_index := range category_proper_tasks{
			var diff = category_proper_tasks[task_index].Deadline.Sub(now_time)
			danger_deadline = append(danger_deadline, diff.Hours() < 5*24)
			rest_day = append(rest_day, int(diff.Hours() / 24 + 1))
		}

		// pagenation
		tasks_length := len(category_proper_tasks)
		str_page_id := ctx.Param("page_id")
		var page_id int 
		if str_page_id != ""{
			// URLパラメータ内にpage_idがある場合
			page_id, err = strconv.Atoi(str_page_id)
			if err != nil {
					Error(http.StatusBadRequest, err.Error())(ctx)
					return
			}
			// セッション内のpage_idを更新
			// セッションの保存
			session := sessions.Default(ctx)
			session.Set("page_id", page_id)
			session.Save()
		}else{
			// URLパラメータ内にpage_idがない場合はセッションを参照
			session_page_id := sessions.Default(ctx).Get("page_id")
			if session_page_id != nil{
				page_id = session_page_id.(int)
			}else{
				page_id = 0
			}
		}

		start_id := page_id * 5
		next_start_id := int(math.Min(float64((page_id + 1) * 5), float64(tasks_length)))
		category_proper_tasks = category_proper_tasks[start_id:next_start_id]
		has_pre_page := page_id > 0
		pre_page_id := page_id - 1
		has_next_page := (tasks_length - start_id) > 5
		next_page_id := page_id + 1

		// // User Information
		// var user_name string
		// err = db.Get(&user_name, "SELECT name FROM users WHERE id = ?",userID.(uint64))
		// if err != nil {
		// 		Error(http.StatusInternalServerError, err.Error())(ctx)
		// 		return
		// }
		user_name := "Foo"

    // Render tasks
    ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": category_proper_tasks, "Kw": kw, "IsDone": str_is_done == "checked", "IsNotDone": str_is_not_done == "checked", "DangerDeadline": danger_deadline, "RestDay": rest_day, "PageId": page_id, "HasPrePage": has_pre_page, "PrePageId": pre_page_id, "HasNextPage": has_next_page, "NextPageId": next_page_id, "UserId": userID, "UserName": user_name})
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
	task_id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	tx := db.MustBegin()
	// Get a task with given ID
	var task database.Task
	err = tx.Get(&task, "SELECT * FROM tasks WHERE id=?", task_id) // Use DB#Get for one entry
	if err != nil {
		tx.Rollback()
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	str_year := strconv.Itoa(task.Deadline.Year())
	str_month := strconv.Itoa(int(task.Deadline.Month()))
	str_day := strconv.Itoa(task.Deadline.Day())
	str_deadline := str_year + "-"  + str_month + "-" + str_day

	var category_ids []int
	err = tx.Select(&category_ids, "SELECT category_id FROM task_category WHERE task_id = ?", task_id)
	if err != nil {
		tx.Rollback()
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	var category_names []string
	for _, category_id := range category_ids{
		var category_name string
		err = tx.Get(&category_name, "SELECT category_name FROM categories WHERE id = ?", category_id)
		if err != nil {
			tx.Rollback()
			Error(http.StatusBadRequest, err.Error())(ctx)
			return
		}		
		category_names = append(category_names, category_name)
	}
	tx.Commit()
	// Render task
	ctx.HTML(http.StatusOK, "task.html", gin.H{"Task": task, "Deadline": str_deadline, "CategoryNames": category_names})
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
	// Get task deadline
	space_linked_category_names, category_exist := ctx.GetPostForm("category_name")

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
	if category_exist && space_linked_category_names != ""{
		category_names := strings.Split(space_linked_category_names, " ")
		for _, category_name := range category_names{
			var category_id int
			err = tx.Get(&category_id, "SELECT id FROM categories WHERE category_name = ?", category_name)
			if err != nil{
				tx.Rollback()
				Error(http.StatusBadRequest, "improper category name")(ctx)
				return
			}
			_, err = tx.Exec("INSERT INTO task_category (task_id, category_id) VALUES (?, ?)", taskID, category_id)
			if err != nil {
					tx.Rollback()
					Error(http.StatusInternalServerError, err.Error())(ctx)
					return
			}
		}
	}
	tx.Commit()
	// Render status
	ctx.Redirect(http.StatusFound, fmt.Sprintf("/task/%d", taskID))
}

func EditTaskForm(ctx *gin.Context) {
	// ID の取得
	task_id, err := strconv.Atoi(ctx.Param("id"))
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
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", task_id)
	if err != nil {
			Error(http.StatusBadRequest, err.Error())(ctx)
			return
	}


	str_year := strconv.Itoa(task.Deadline.Year())
	str_month := strconv.Itoa(int(task.Deadline.Month()))
	str_day := strconv.Itoa(task.Deadline.Day())
	str_deadline := str_year + "-"  + str_month + "-" + str_day

	tx := db.MustBegin()
	var category_ids []int
	err = tx.Select(&category_ids, "SELECT category_id FROM task_category WHERE task_id = ?", task_id)
	if err != nil {
		tx.Rollback()
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	var category_names []string
	for _, category_id := range category_ids{
		var category_name string
		err = tx.Get(&category_name, "SELECT category_name FROM categories WHERE id = ?", category_id)
		if err != nil {
			tx.Rollback()
			Error(http.StatusBadRequest, err.Error())(ctx)
			return
		}		
		category_names = append(category_names, category_name)
	}
	tx.Commit()

	// Render edit form
	ctx.HTML(http.StatusOK, "form_edit_task.html",
			gin.H{"Title": fmt.Sprintf("Edit task %d", task.ID), "Task": task, "Deadline": str_deadline, "CategoryName": strings.Join(category_names, " ")})
}

func UpdateTask(ctx *gin.Context){
	// ID の取得
	task_id, err := strconv.Atoi(ctx.Param("id"))
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
	// Get task deadline
	deadline, exist := ctx.GetPostForm("deadline")
	if !exist {
			Error(http.StatusBadRequest, "No deadline is given")(ctx)
			return
	}
	// Get task category
	space_linked_category_names, category_exist := ctx.GetPostForm("category_name")

	// Get share id
	space_linked_user_ids, to_share_user_ids_exist := ctx.GetPostForm("to_share_user_ids")

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}
	tx := db.MustBegin()
	// Create new data with given title on DB
	tx.Exec("UPDATE tasks SET title = ?, comment = ?, is_done = ?, priority = ?, deadline = ? WHERE id = ?", title, comment, bool_is_done, int_priority, deadline, task_id)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Delete the task_category from DB
	_, err = tx.Exec("DELETE FROM task_category WHERE task_id=?", task_id)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	if category_exist && space_linked_category_names != ""{
		category_names := strings.Split(space_linked_category_names, " ")
		for _, category_name := range category_names{
			var category_id int
			err = tx.Get(&category_id, "SELECT id FROM categories WHERE category_name = ?", category_name)
			if err != nil{
				tx.Rollback()
				Error(http.StatusBadRequest, "improper category name")(ctx)
				return
			}
			_, err = tx.Exec("INSERT INTO task_category (task_id, category_id) VALUES (?, ?)", task_id, category_id)
			if err != nil {
				tx.Rollback()
				Error(http.StatusInternalServerError, err.Error())(ctx)
				return
			}
		}
	}

	if to_share_user_ids_exist && space_linked_user_ids != ""{
		user_ids := strings.Split(space_linked_user_ids, " ")
		for _, user_id := range user_ids{
			_, err = tx.Exec("INSERT INTO ownerships (task_id, user_id) VALUES (?, ?)", task_id, user_id)
			if err != nil {
				tx.Rollback()
				Error(http.StatusInternalServerError, err.Error())(ctx)
				return
			}
		}
	}
	tx.Commit()
	// Render status
	path := fmt.Sprintf("/task/%d", task_id)
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
	ctx.Redirect(http.StatusFound, "/list/0")
}

// func ShareTask(ctx *gin.Context){
// 	task_id := ctx.Param("task_id")
// 	space_linked_user_ids, exist := ctx.GetPostForm("to_share_user_ids")

// 	// Get DB connection
// 	db, err := database.GetConnection()
// 	if err != nil {
// 			Error(http.StatusInternalServerError, err.Error())(ctx)
// 			return
// 	}
// 	if exist {
// 		user_ids := strings.Split(space_linked_user_ids, " ")
// 		tx := db.MustBegin()
// 		for user_id := range user_ids{
// 			_, err = tx.Exec("INSERT INTO ownerships (task_id, user_id) VALUES (?, ?)", task_id, user_id)
// 			if err != nil {
// 				tx.Rollback()
// 				Error(http.StatusInternalServerError, err.Error())(ctx)
// 				return
// 			}
// 		}
// 		tx.Commit()
// 	}
// 	path := "/task/" + task_id
// 	ctx.Redirect(http.StatusFound, path)
// }
