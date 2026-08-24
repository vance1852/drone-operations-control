package domain

func StatusCounts(drone_tasks []DroneTask) map[DroneTaskStatus]int {
	counts := make(map[DroneTaskStatus]int)
	for _, task := range drone_tasks {
		counts[task.Status]++
	}
	return counts
}
