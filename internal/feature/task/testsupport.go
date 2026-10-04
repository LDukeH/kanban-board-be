package task

func BuildTaskUpdateSet(task *Task) ([]string, []any) {
	return buildTaskUpdateSet(task)
}

func IsValidTaskPriority(p string) bool {
	return isValidTaskPriority(p)
}
