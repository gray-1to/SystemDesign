package db

// schema.go provides data models in DB
import (
	"time"
)

// Task corresponds to a row in `tasks` table
type Task struct {
	ID        uint64    `db:"id"`
	Title     string    `db:"title"`
	Comment   string    `db:"comment"`
	CreatedAt time.Time `db:"created_at"`
	IsDone    bool      `db:"is_done"`
	Priority  uint64    `db:"priority"`
	Deadline  time.Time `db:"deadline"`
}

type User struct {
	ID        uint64    `db:"id"`
	Name      string    `db:"name"`
	Password  []byte    `db:"password"`
}

type Ownership struct {
	User_id uint64 `db:"user_id"`
	Task_id uint64 `db:"task_id"`
}

type Category struct {
	ID        uint64   `db:"id"`
	CategoryName string `db:"category_name"`
}

type TaskCategory struct {
	Task_id  uint `db:"task_id"`
	Category_id uint `db:"category_id"`
}
