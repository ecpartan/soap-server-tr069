package dao

type Task struct {
	Id         string `db:"id"`
	Name       string `db:"name"`
	LastStatus string `db:"last_status"`
	EventCode  int    `db:"event_code"`
	Once       bool   `db:"once"`
	TaskOpId   string `db:"task_op_id"`
	DeviceId   string `db:"device_id"`
}

type TaskResult struct {
	Id        string `db:"id"`
	Status    string `db:"status"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
	Result    string `db:"result"`
	TaskId    string `db:"task_id"`
}

type TaskOp struct {
	Id   string `db:"id"`
	Name string `db:"name"`
	Body string `db:"body"`
}
