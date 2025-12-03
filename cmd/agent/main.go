package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Mukam21/ai-browser-agent/internal/agent"
	"github.com/Mukam21/ai-browser-agent/internal/browser"
	"github.com/Mukam21/ai-browser-agent/internal/config"
	"github.com/Mukam21/ai-browser-agent/internal/llm"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx := context.Background()
	cfg := config.Load()

	showHeader()

	if cfg.OpenAIAPIKey == "" {
		showAPIKeyHelp()
		return
	}

	log.Println(" Конфигурация загружена:")
	log.Printf("  • Модель: %s", cfg.Model)
	log.Printf("  • Макс. шагов: %d", cfg.MaxSteps)
	log.Printf("  • Режим отладки: %v", cfg.Debug)

	log.Println("🧠 Инициализирую LLM клиент...")
	llmClient := llm.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.Model, cfg.MaxTokens)

	log.Println("🚀 Инициализирую браузер...")
	browserCtrl, err := browser.NewController(ctx)
	if err != nil {
		log.Fatalf("❌ Ошибка инициализации браузера: %v", err)
	}
	defer browserCtrl.Close()

	log.Println("✅ Система инициализирована успешно!")

	agentConfig := &agent.Config{
		MaxSteps:      cfg.MaxSteps,
		Debug:         cfg.Debug,
		Model:         cfg.Model,
		MaxMemorySize: 50,
		StepDelay:     2 * time.Second,
		Timeout:       time.Duration(cfg.Timeout) * time.Second,
	}

	autonomousAgent := agent.NewAutonomousAgent(llmClient, browserCtrl, agentConfig)

	runMainLoop(ctx, autonomousAgent)
}

func showHeader() {
	clearScreen()

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("🤖 AI BROWSER AGENT - Автономный агент для управления браузером")
	fmt.Println("🚀 Версия 1.0.0 | Архитектура: ReAct (Reasoning + Acting)")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println()
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func showAPIKeyHelp() {
	fmt.Println(`
❌ OPENAI_API_KEY не установлен!

Для работы AI агента требуется API ключ OpenAI.

📋 Как получить ключ:
1. Попросите у рекрутера (рекомендуется)
2. Или создайте на platform.openai.com

🔧 Как установить ключ:

Linux/Mac:
export OPENAI_API_KEY="sk-ваш_ключ"

Windows (CMD):
set OPENAI_API_KEY=sk-ваш_ключ

Windows (PowerShell):
$env:OPENAI_API_KEY="sk-ваш_ключ"

🌐 Дополнительные настройки в config/config.go`)

	fmt.Print("\nНажмите Enter для выхода...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func runMainLoop(ctx context.Context, agent *agent.AutonomousAgent) {
	for {
		showMenu()

		choice := getUserChoice()

		switch choice {
		case "1":
			executeTask(ctx, agent)
		case "2":
			showExamples()
		case "3":
			showCapabilities()
		case "4":
			fmt.Println("\n👋 До свидания!")
			return
		default:
			fmt.Println("\n❌ Неверный выбор. Попробуйте снова.")
		}
	}
}

func showMenu() {
	fmt.Println("\n" + strings.Repeat("─", 70))
	fmt.Println("ГЛАВНОЕ МЕНЮ")
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println("1. Выполнить задачу")
	fmt.Println("2. Примеры задач")
	fmt.Println("3. Возможности агента")
	fmt.Println("4. Выход")
	fmt.Println(strings.Repeat("─", 70))

	fmt.Print("\n👉 Выберите действие (1-4): ")
}

func getUserChoice() string {
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	return strings.TrimSpace(choice)
}

func executeTask(ctx context.Context, agent *agent.AutonomousAgent) {
	fmt.Println("\n" + strings.Repeat("═", 70))
	fmt.Println("ВЫПОЛНЕНИЕ ЗАДАЧИ")
	fmt.Println(strings.Repeat("═", 70))

	fmt.Println("\n Введите задачу для AI агента:")
	fmt.Println("(Пример: 'Найди информацию про искусственный интеллект на Википедии')")

	fmt.Print("\n👉 Задача: ")

	reader := bufio.NewReader(os.Stdin)
	task, _ := reader.ReadString('\n')
	task = strings.TrimSpace(task)

	if task == "" {
		fmt.Println("\n Задача не может быть пустой.")
		return
	}

	fmt.Printf("\n🧠 Анализирую задачу: \"%s\"\n", task)
	fmt.Println(" AI агент начинает выполнение...")
	fmt.Println(strings.Repeat("─", 70))

	startTime := time.Now()
	result, err := agent.Execute(ctx, task)
	elapsed := time.Since(startTime)

	fmt.Println("\n" + strings.Repeat("═", 70))
	fmt.Println(" РЕЗУЛЬТАТ ВЫПОЛНЕНИЯ")
	fmt.Println(strings.Repeat("═", 70))

	if err != nil {
		fmt.Printf("❌ Ошибка выполнения: %v\n", err)
	} else {
		fmt.Println(result)
	}

	fmt.Printf("\n Время выполнения: %v\n", elapsed.Round(time.Second))
	fmt.Println(strings.Repeat("═", 70))

	fmt.Print("\nНажмите Enter для продолжения...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func showExamples() {
	examples := `
📚 ПРИМЕРЫ ЗАДАЧ ДЛЯ AI АГЕНТА

1.  Поиск информации:
   • "Найди последние новости про искусственный интеллект"
   • "Найди информацию о Go программировании на официальном сайте"
   • "Поищи курсы по машинному обучению"

2.  Навигация и исследование:
   • "Перейди на сайт GitHub и найди trending репозитории"
   • "Открой Википедию и найди статью про квантовые компьютеры"
   • "Исследуй документацию Docker"

3.  Работа с почтой (концепт):
   • "Проверь новые письма в почте" (требует авторизацию)
   • "Найди письма от конкретного отправителя"

4.  Онлайн-покупки (концепт):
   • "Найди товары по определенной категории"
   • "Сравни цены на продукт"

5.  Поиск работы (концепт):
   • "Найди вакансии Go разработчика"
   • "Ищи удаленные позиции"

 Примечание: Некоторые задачи требуют авторизации или 
   дополнительной настройки безопасности.
`

	fmt.Println(examples)
	fmt.Print("\nНажмите Enter для продолжения...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func showCapabilities() {
	capabilities := `
🔧 ВОЗМОЖНОСТИ AI АГЕНТА

✅ Автономное выполнение задач:
   • Агент сам планирует действия через LLM
   • Анализирует страницы в реальном времени
   • Принимает решения на основе контекста

✅ Безопасная архитектура:
   • ReAct паттерн (Reasoning + Acting)
   • Управление контекстом и памятью
   • Обработка ошибок и восстановление

✅ Поддержка различных сценариев:
   • Поиск и сбор информации
   • Навигация по сложным сайтам
   • Взаимодействие с формами
   • Многошаговые workflows

✅ Технические особенности:
   • Интеграция с OpenAI GPT
   • Расширяемая архитектура инструментов
   • Векторная память для долгосрочного контекста
   • Поддержка разных браузерных движков

 Архитектурные принципы:
1. Нет предзаданных шагов - агент исследует сайты сам
2. Нет хардкода селекторов - использует описания
3. Контекстное управление - помнит историю действий
4. Self-correction - исправляет ошибки автоматически

 Планы развития:
• Компьютерное зрение для анализа интерфейсов
• Поддержка мульти-агентных систем
• Интеграция с внешними API
• Веб-интерфейс для управления
`

	fmt.Println(capabilities)
	fmt.Print("\nНажмите Enter для продолжения...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
