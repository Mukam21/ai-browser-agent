package agent

import (
	"fmt"
	"time"
)

type TaskContext struct {
	TaskID       string
	OriginalTask string
	CurrentGoal  string
	Progress     []ProgressStep
	CurrentURL   string
	PageHistory  []PageState
	StartTime    time.Time
}

type ProgressStep struct {
	StepNumber int
	Goal       string
	Action     string
	Result     string
	Timestamp  time.Time
	Success    bool
}

type PageState struct {
	URL       string
	Title     string
	Timestamp time.Time
	Summary   string
}

func NewTaskContext(task string) *TaskContext {
	return &TaskContext{
		TaskID:       generateTaskID(),
		OriginalTask: task,
		CurrentGoal:  "Начать выполнение задачи",
		Progress:     make([]ProgressStep, 0),
		PageHistory:  make([]PageState, 0),
		StartTime:    time.Now(),
	}
}

func (tc *TaskContext) AddProgress(stepNum int, goal, action, result string, success bool) {
	step := ProgressStep{
		StepNumber: stepNum,
		Goal:       goal,
		Action:     action,
		Result:     result,
		Timestamp:  time.Now(),
		Success:    success,
	}

	tc.Progress = append(tc.Progress, step)
	tc.CurrentGoal = goal
}

func (tc *TaskContext) AddPageState(url, title, summary string) {
	state := PageState{
		URL:       url,
		Title:     title,
		Timestamp: time.Now(),
		Summary:   summary,
	}

	tc.PageHistory = append(tc.PageHistory, state)
	tc.CurrentURL = url
}

func (tc *TaskContext) GetProgressSummary() string {
	if len(tc.Progress) == 0 {
		return "Прогресс: еще не начато"
	}

	completed := 0
	for _, step := range tc.Progress {
		if step.Success {
			completed++
		}
	}

	return fmt.Sprintf("Прогресс: %d/%d шагов завершено", completed, len(tc.Progress))
}

func (tc *TaskContext) GetRecentPageStates(count int) string {
	if len(tc.PageHistory) == 0 {
		return "История страниц: пусто"
	}

	start := len(tc.PageHistory) - count
	if start < 0 {
		start = 0
	}

	recent := tc.PageHistory[start:]
	result := "📄 Недавние страницы:\n"

	for i, state := range recent {
		result += fmt.Sprintf("%d. %s - %s\n", i+1, state.Title, state.URL)
	}

	return result
}

func (tc *TaskContext) GetCurrentContext() string {
	context := fmt.Sprintf(`🎯 Текущая задача: %s
📈 %s
🎯 Текущая цель: %s
🌐 Текущий URL: %s

`, tc.OriginalTask, tc.GetProgressSummary(), tc.CurrentGoal, tc.CurrentURL)

	// Добавляем последние шаги прогресса
	if len(tc.Progress) > 0 {
		context += "📋 Последние действия:\n"
		start := len(tc.Progress) - 3
		if start < 0 {
			start = 0
		}

		for i := start; i < len(tc.Progress); i++ {
			step := tc.Progress[i]
			status := "✅"
			if !step.Success {
				status = "❌"
			}
			context += fmt.Sprintf("%s Шаг %d: %s -> %s\n",
				status, step.StepNumber, step.Goal, truncateText(step.Result, 80))
		}
		context += "\n"
	}

	return context
}

func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().Unix())
}
