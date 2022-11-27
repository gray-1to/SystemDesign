package service
 
import (
		"crypto/sha256"
    "net/http"
		"unicode/utf8"
		"encoding/hex"
		"strconv"
		"fmt"
 
    "github.com/gin-gonic/gin"
		"github.com/gin-contrib/sessions"
		database "todolist.go/db"
)
 
func NewUserForm(ctx *gin.Context) {
    ctx.HTML(http.StatusOK, "new_user_form.html", gin.H{"Title": "Register user"})
}

func hash(pw string) []byte {
	const salt = "todolist.go#"
	h := sha256.New()
	h.Write([]byte(salt))
	h.Write([]byte(pw))
	return h.Sum(nil)
}

func RegisterUser(ctx *gin.Context) {
	// フォームデータの受け取り
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	password_confirm := ctx.PostForm("password_confirm")
	switch {
	case username == "":
			ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is not provided", "Username": username, "Password": password, "PasswordConfirm": password_confirm})
			return
	case password == "":
			ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is not provided", "Username": username, "Password": password, "PasswordConfirm": password_confirm})
			return
	case password != password_confirm:
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password confirmation is not same ", "Username": username, "Password": password, "PasswordConfirm": password_confirm})
		return
	case utf8.RuneCountInString(password) < 6:
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is too short. Set Password more than 5 chars.", "Username": username, "Password": password, "PasswordConfirm": password_confirm})
		return
	}
	
	// DB 接続
	db, err := database.GetConnection()
	if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}
 
	// 重複チェック
	var duplicate int
	err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username)
	if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}
	if duplicate > 0 {
			ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "Username": username, "Password": password})
			return
	}

	// DB への保存
	result, err := db.Exec("INSERT INTO users(name, password) VALUES (?, ?)", username, hash(password))
	if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}

	// 保存状態の確認
	id, _ := result.LastInsertId()
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", id)
	if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}

	// セッションの保存
	session := sessions.Default(ctx)
	session.Set(userkey, user.ID)
	session.Save()

	TaskList(ctx)
}

const userkey = "user"

func ShowLoginForm(ctx *gin.Context){
	ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login"})
}
 
func Login(ctx *gin.Context) {
    username := ctx.PostForm("username")
    password := ctx.PostForm("password")
 
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
 
    // ユーザの取得
    var user database.User
    err = db.Get(&user, "SELECT id, name, password FROM users WHERE name = ?", username)
    if err != nil {
        ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "No such user"})
        return
    }
 
    // パスワードの照合
    if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
        ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "Incorrect password"})
        return
    }
 
    // セッションの保存
    session := sessions.Default(ctx)
    session.Set(userkey, user.ID)
    session.Save()
 
    ctx.Redirect(http.StatusFound, "/list")
}

func LoginCheck(ctx *gin.Context) {
	user_id := sessions.Default(ctx).Get(userkey)
	if user_id == nil {
			ctx.Redirect(http.StatusFound, "/login")
			ctx.Abort()
	} else {
		// ID の取得
		str_task_id := ctx.Param("id")
		if str_task_id == "" {
			// taskのidが指定されてないもの
			ctx.Next()
		}else{
			int_task_id, err := strconv.Atoi(str_task_id)
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

			var ownerships []database.Ownership
			db.Select(&ownerships, "SELECT * FROM ownerships WHERE user_id = ? AND task_id = ?", user_id, int_task_id)

			if len(ownerships) == 0{
				ctx.JSON(http.StatusNotFound, fmt.Sprintf("you do not have access right. task_id: %d, user_id: %d", int_task_id, user_id))
				ctx.Abort()
			}else{
				ctx.Next()
			}
		}
			
	}
}

func Logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	session.Save()
	ctx.Redirect(http.StatusFound, "/")
}

func DeleteUser(ctx *gin.Context){
	// ID の取得
	id := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
	}
	tx := db.MustBegin()
	// Delete the tasks from DB
	_, err = tx.Exec("DELETE tasks FROM tasks INNER JOIN ownerships ON ownerships.task_id = tasks.id WHERE ownerships.user_id = ?", id)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Delete the ownerships from DB
	_, err = tx.Exec("DELETE FROM ownerships WHERE user_id=?", id)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Delete the user from DB
	_, err = tx.Exec("DELETE FROM users WHERE id=?", id)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	tx.Commit()

	session := sessions.Default(ctx)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	session.Save()
	// Redirect to /list
	ctx.Redirect(http.StatusFound, "/")
}
