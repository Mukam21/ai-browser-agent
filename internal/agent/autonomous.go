package agent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Mukam21/ai-browser-agent/internal/browser"
	"github.com/Mukam21/ai-browser-agent/internal/llm"
)

type AutonomousAgent struct {
	llm       llm.Client
	browser   browser.Controller
	planner   *Planner
	memory    *Memory
	context   *TaskContext
	config    *Config
	isRunning bool
}

type Config struct {
	MaxSteps      int
	Debug         bool
	Model         string
	MaxMemorySize int
	StepDelay     time.Duration
	Timeout       time.Duration
}

func NewAutonomousAgent(llm llm.Client, browser browser.Controller, config *Config) *AutonomousAgent {
	if config == nil {
		config = &Config{
			MaxSteps:      30,
			Debug:         true,
			Model:         "gpt-4o-mini",
			MaxMemorySize: 50,
			StepDelay:     2 * time.Second,
			Timeout:       5 * time.Minute,
		}
	}

	memory := NewMemory(config.MaxMemorySize)
	context := NewTaskContext("")

	return &AutonomousAgent{
		llm:     llm,
		browser: browser,
		memory:  memory,
		context: context,
		config:  config,
		planner: NewPlanner(llm, memory, context),
	}
}

func (a *AutonomousAgent) Execute(ctx context.Context, task string) (string, error) {
	if a.isRunning {
		return "", fmt.Errorf("агент уже выполняет задачу")
	}

	a.isRunning = true
	defer func() { a.isRunning = false }()

	timeoutCtx, cancel := context.WithTimeout(ctx, a.config.Timeout)
	defer cancel()

	log.Printf(" Начинаю выполнение задачи: %s", task)

	a.context = NewTaskContext(task)
	a.planner = NewPlanner(a.llm, a.memory, a.context)

	for stepNum := 1; stepNum <= a.config.MaxSteps; stepNum++ {
		select {
		case <-timeoutCtx.Done():
			return "", fmt.Errorf("таймаут выполнения задачи")
		default:
		}

		log.Printf("\n Шаг %d/%d", stepNum, a.config.MaxSteps)

		observation, err := a.observe(timeoutCtx)
		if err != nil {
			log.Printf(" Ошибка наблюдения: %v", err)
			observation = fmt.Sprintf("Ошибка получения состояния: %v", err)
		}

		plan, err := a.planner.PlanNextAction(timeoutCtx, observation)
		if err != nil {
			log.Printf("❌ Ошибка планирования: %v", err)
			return "", fmt.Errorf("ошибка планирования: %w", err)
		}

		log.Printf(" Мысль: %s", plan.Thought)
		log.Printf("Действие: %s", plan.Action)
		if a.config.Debug {
			log.Printf("Reasoning: %s", plan.Reasoning)
			log.Printf(" Confidence: %.2f", plan.Confidence)
		}

		result, err := a.act(timeoutCtx, plan)
		if err != nil {
			result = fmt.Sprintf("Ошибка: %v", err)
			log.Printf(" Ошибка выполнения: %v", err)
		} else {
			log.Printf("✅ Результат: %s", truncateText(result, 100))
		}

		success := err == nil && !strings.Contains(strings.ToLower(result), "ошибка")
		a.planner.EvaluateResult(plan, result, success)

		if plan.Action == "finish" {
			log.Println("🎉 Задача выполнена успешно!")
			return a.generateFinalReport(plan, result), nil
		}

		time.Sleep(a.config.StepDelay)
	}

	return a.generateTimeoutReport(), nil
}

func (a *AutonomousAgent) observe(ctx context.Context) (string, error) {
	state, err := a.browser.GetPageState()
	if err != nil {
		return "", err
	}

	dom, domErr := a.browser.GetDOM()

	observation := fmt.Sprintf(`Текущее состояние страницы:
%s

Дополнительная информация:
- Модель LLM: %s
- Память агента: %d взаимодействий
- Текущий шаг: %d/%d`,
		state, a.llm.GetModel(), len(a.memory.interactions),
		len(a.context.Progress)+1, a.config.MaxSteps)

	if domErr == nil {
		domPreview := truncateText(dom, 2000)
		observation += fmt.Sprintf("\n\nHTML (первые 2000 символов):\n%s", domPreview)
	}

	a.context.AddPageState("current", "Текущая страница", truncateText(state, 100))

	return observation, nil
}

func (a *AutonomousAgent) act(ctx context.Context, plan *Plan) (string, error) {
	switch plan.Action {
	case "navigate":
		url, _ := plan.ActionInput["url"].(string)
		if err := a.browser.Navigate(url); err != nil {
			return "", fmt.Errorf("ошибка навигации: %w", err)
		}
		return fmt.Sprintf("Успешно перешел на %s", url), nil

	case "find_element":
		description, _ := plan.ActionInput["description"].(string)
		return fmt.Sprintf("Найден элемент по описанию: '%s'", description), nil

	case "click":
		description, _ := plan.ActionInput["description"].(string)
		if err := a.browser.Click(description); err != nil {
			return "", fmt.Errorf("ошибка клика: %w", err)
		}
		return fmt.Sprintf("Успешно кликнул на элемент: '%s'", description), nil

	case "type":
		description, _ := plan.ActionInput["description"].(string)
		text, _ := plan.ActionInput["text"].(string)
		if err := a.browser.Fill(description, text); err != nil {
			return "", fmt.Errorf("ошибка ввода текста: %w", err)
		}
		return fmt.Sprintf("Успешно ввел текст '%s' в элемент: '%s'", text, description), nil

	case "read":
		state, err := a.browser.GetPageState()
		if err != nil {
			return "", fmt.Errorf("ошибка чтения: %w", err)
		}
		return fmt.Sprintf("Прочитано состояние страницы: %s", truncateText(state, 150)), nil

	case "scroll":
		return "Страница прокручена", nil

	case "wait":
		seconds, _ := plan.ActionInput["seconds"].(float64)
		duration := time.Duration(seconds) * time.Second
		time.Sleep(duration)
		return fmt.Sprintf("Ожидание завершено (%.0f секунд)", seconds), nil

	case "screenshot":
		filename := fmt.Sprintf("screenshot_step_%d_%d.png",
			len(a.context.Progress)+1, time.Now().Unix())
		if err := a.browser.Screenshot(filename); err != nil {
			return "", fmt.Errorf("ошибка скриншота: %w", err)
		}
		return fmt.Sprintf("Скриншот сохранен: %s", filename), nil

	case "finish":
		result, _ := plan.ActionInput["result"].(string)
		return fmt.Sprintf("Задача завершена: %s", result), nil

	default:
		return "", fmt.Errorf("неизвестное действие: %s", plan.Action)
	}
}

func (a *AutonomousAgent) generateFinalReport(plan *Plan, result string) string {
	finalResult, _ := plan.ActionInput["result"].(string)

	report := fmt.Sprintf(` ЗАДАЧА ВЫПОЛНЕНА УСПЕШНО!

 Исходная задача: %s

 Статистика выполнения:
• Всего шагов: %d
• Успешных шагов: %d
• Время выполнения: %v
• Использованная модель: %s

 Результат: %s

 История действий:`,
		a.context.OriginalTask,
		len(a.context.Progress),
		a.countSuccessfulSteps(),
		time.Since(a.context.StartTime).Round(time.Second),
		a.llm.GetModel(),
		finalResult)

	for _, step := range a.context.Progress {
		status := "✅"
		if !step.Success {
			status = "❌"
		}
		report += fmt.Sprintf("\n%s Шаг %d: %s -> %s",
			status, step.StepNumber, step.Goal, truncateText(step.Result, 80))
	}

	report += "\n\n Insights из памяти агента:"
	insights := a.memory.GetInsights()
	if len(insights) == 0 {
		report += "\nНет значимых insights"
	} else {
		for key, insight := range insights {
			report += fmt.Sprintf("\n- %s: %s", key, insight)
		}
	}

	return report
}

func (a *AutonomousAgent) generateTimeoutReport() string {
	report := fmt.Sprintf(` ДОСТИГНУТ ЛИМИТ ШАГОВ

 Исходная задача: %s

 Статистика выполнения:
• Выполнено шагов: %d
• Успешных шагов: %d
• Максимальный лимит: %d шагов
• Время выполнения: %v

 Причина остановки: достигнут максимальный лимит шагов

 История последних действий:`,
		a.context.OriginalTask,
		len(a.context.Progress),
		a.countSuccessfulSteps(),
		a.config.MaxSteps,
		time.Since(a.context.StartTime).Round(time.Second))

	start := len(a.context.Progress) - 5
	if start < 0 {
		start = 0
	}

	for idx := start; idx < len(a.context.Progress); idx++ {
		step := a.context.Progress[idx]
		status := "✅"
		if !step.Success {
			status = "❌"
		}
		report += fmt.Sprintf("\n%s Шаг %d: %s -> %s",
			status, step.StepNumber, step.Goal, truncateText(step.Result, 80))
	}

	report += "\n\n Рекомендации:"
	report += "\n• Упростите задачу или разбейте на подзадачи"
	report += "\n• Увеличьте MAX_STEPS в конфигурации"
	report += "\n• Проверьте доступность целевых сайтов"

	return report
}

func (a *AutonomousAgent) countSuccessfulSteps() int {
	count := 0
	for _, step := range a.context.Progress {
		if step.Success {
			count++
		}
	}
	return count
}

func (a *AutonomousAgent) Stop() {
	a.isRunning = false
	if a.browser != nil {
		a.browser.Close()
	}
}
