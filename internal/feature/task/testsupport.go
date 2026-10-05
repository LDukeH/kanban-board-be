package task

func BuildTaskUpdateSet(task *Task) ([]string, []any) {
	return buildTaskUpdateSet(task)
}

func IsValidTaskPriority(p string) bool {
	return isValidTaskPriority(p)
}

// ClassifyVersionMismatch exposes the mapping for tests.
func ClassifyVersionMismatch(exists bool) error {
	return classifyVersionMismatch(exists)
}
