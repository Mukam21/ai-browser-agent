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

type Plan struct {
	Thought     string                 `json:"thought"`
	Action      string                 `json:"action"`
	ActionInput map[string]interface{} `json:"action_input"`
	Reasoning   string                 `json:"reasoning"`
	Confidence  float64                `json:"confidence"`
}

func NewPlanner(llm llm.Client, memory *Memory, context *TaskContext) *Planner {
	return &Planner{
		llm:     llm,
		memory:  memory,
		context: context,
	}
}

func (p *Planner) PlanNextAction(ctx context.Context, observation string) (*Plan, error) {
	prompt := p.buildPlanningPrompt(observation)

	var plan Plan
	if err := p.llm.ChatStructured(ctx, prompt, &plan); err != nil {
		return nil, fmt.Errorf("ошибка планирования: %w", err)
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
