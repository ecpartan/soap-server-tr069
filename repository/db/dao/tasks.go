package dao

type WebTask struct {
	Id          int    `db:"id"`
	TaskId      string `db:"task_id"`
	TaskName    string `db:"task_name"`
	TaskType    string `db:"task_type"`
	TaskStatus  string `db:"task_status"`
	TaskContent string `db:"task_content"`
	CreateTime  string `db:"create_time"`
	UpdateTime  string `db:"update_time"`
}
