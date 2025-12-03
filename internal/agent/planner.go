package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Mukam21/ai-browser-agent/internal/llm"
)

type Planner struct {
	llm     llm.Client
	memory  *Memory
	context *TaskContext
}

func NewPlanner(llm llm.Client, memory *Memory, context *TaskContext) *Planner {
	return &Planner{
		llm:     llm,
		memory:  memory,
		context: context,
	}
}

type Plan struct {
	Thought     string                 `json:"thought"`
	Action      string                 `json:"action"`
	ActionInput map[string]interface{} `json:"action_input"`
	Reasoning   string                 `json:"reasoning"`
	Confidence  float64                `json:"confidence"`
}

func (p *Planner) PlanNextAction(ctx context.Context, observation string) (*Plan, error) {
	prompt := p.buildPlanningPrompt(observation)

	var plan Plan
	if err := p.llm.ChatStructured(ctx, prompt, &plan); err != nil {
		return nil, fmt.Errorf("ошибка планирования: %w", err)
	}

	// Валидация плана
	if !p.validatePlan(&plan) {
		return p.createFallbackPlan(), nil
	}

	return &plan, nil
}

func (p *Planner) buildPlanningPrompt(observation string) string {
	memoryContext := p.memory.GetContext()
	taskContext := p.context.GetCurrentContext()

	prompt := fmt.Sprintf(`Ты - автономный AI агент для управления веб-браузером. Твоя задача: "%s"

%s

%s

Текущее состояние страницы:
%s

Доступные действия:
1. navigate - перейти на URL (требует параметр "url")
2. find_element - найти элемент по описанию (требует "description")
3. click - кликнуть на элемент (требует "description" элемента)
4. type - ввести текст (требует "description" элемента и "text")
5. read - прочитать информацию со страницы
6. scroll - прокрутить страницу вниз
7. wait - подождать N секунд (требует "seconds")
8. screenshot - сделать скриншот
9. finish - завершить задачу (требует "result" - описание результата)

Критические правила:
1. НИКОГДА не используй заранее известные селекторы или структуры конкретных сайтов
2. ВСЕГДА анализируй текущую страницу, чтобы понять что делать
3. Если нужно найти элемент, опиши его в "description" (например: "кнопка поиска", "поле для ввода email")
4. Сначала найди элемент (find_element), потом взаимодействуй с ним (click/type)
5. Если задача выполнена, используй действие "finish"

Верни ответ в формате JSON:
{
    "thought": "Твои рассуждения о том, что делать дальше",
    "action": "название действия из списка выше",
    "action_input": {"параметры": "для действия"},
    "reasoning": "подробное объяснение почему выбрано это действие",
    "confidence": 0.95
}`, p.context.OriginalTask, taskContext, memoryContext, observation)

	return prompt
}

func (p *Planner) validatePlan(plan *Plan) bool {
	if plan.Action == "" {
		return false
	}

	validActions := map[string]bool{
		"navigate":     true,
		"find_element": true,
		"click":        true,
		"type":         true,
		"read":         true,
		"scroll":       true,
		"wait":         true,
		"screenshot":   true,
		"finish":       true,
	}

	if !validActions[plan.Action] {
		return false
	}

	// Проверяем необходимые параметры для каждого действия
	switch plan.Action {
	case "navigate":
		if url, ok := plan.ActionInput["url"].(string); !ok || url == "" {
			return false
		}
	case "type":
		if _, ok := plan.ActionInput["description"].(string); !ok {
			return false
		}
		if _, ok := plan.ActionInput["text"].(string); !ok {
			return false
		}
	case "wait":
		if seconds, ok := plan.ActionInput["seconds"].(float64); !ok || seconds <= 0 {
			return false
		}
	case "finish":
		if _, ok := plan.ActionInput["result"].(string); !ok {
			return false
		}
	}

	return true
}

func (p *Planner) createFallbackPlan() *Plan {
	return &Plan{
		Thought:     "Попробую прочитать информацию со страницы чтобы понять что делать дальше",
		Action:      "read",
		ActionInput: map[string]interface{}{},
		Reasoning:   "Fallback plan: нужно больше информации о текущей странице",
		Confidence:  0.5,
	}
}

func (p *Planner) EvaluateResult(plan *Plan, result string, success bool) {
	interaction := Interaction{
		Timestamp: time.Now(),
		Action:    plan.Action,
		Input:     plan.ActionInput,
		Result:    result,
		Important: !success || p.isImportantResult(result),
	}

	p.memory.AddInteraction(interaction)

	goal := fmt.Sprintf("%s: %s", plan.Action, truncateText(plan.Thought, 50))
	p.context.AddProgress(
		len(p.context.Progress)+1,
		goal,
		plan.Action,
		result,
		success,
	)
}

func (p *Planner) isImportantResult(result string) bool {
	keywords := []string{
		"успех", "ошибка", "найден", "не найден",
		"завершено", "провал", "критический",
	}

	resultLower := strings.ToLower(result)
	for _, keyword := range keywords {
		if strings.Contains(resultLower, keyword) {
			return true
		}
	}

	return false
}
