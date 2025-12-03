package agent

import (
	"fmt"
	"strings"
	"time"
)

type Memory struct {
	interactions []Interaction
	insights     map[string]string
	maxSize      int
}

type Interaction struct {
	Timestamp time.Time
	Action    string
	Input     map[string]interface{}
	Result    string
	PageState string
	Important bool
}

func NewMemory(maxSize int) *Memory {
	return &Memory{
		interactions: make([]Interaction, 0),
		insights:     make(map[string]string),
		maxSize:      maxSize,
	}
}

func (m *Memory) AddInteraction(interaction Interaction) {
	m.interactions = append(m.interactions, interaction)

	if len(m.interactions) > m.maxSize {
		m.interactions = m.interactions[1:]
	}

	if interaction.Important {
		m.extractInsights(interaction)
	}
}

func (m *Memory) extractInsights(interaction Interaction) {
	key := interaction.Action

	if strings.Contains(strings.ToLower(interaction.Result), "успех") ||
		strings.Contains(strings.ToLower(interaction.Result), "найден") ||
		strings.Contains(strings.ToLower(interaction.Result), "работает") {
		m.insights[key] = "Это действие работает хорошо"
	}

	if strings.Contains(strings.ToLower(interaction.Result), "ошибка") ||
		strings.Contains(strings.ToLower(interaction.Result), "не найден") ||
		strings.Contains(strings.ToLower(interaction.Result), "не работает") {
		m.insights[key] = "Это действие требует доработки"
	}
}

func (m *Memory) GetRecentInteractions(count int) []Interaction {
	if count > len(m.interactions) {
		count = len(m.interactions)
	}

	start := len(m.interactions) - count
	if start < 0 {
		start = 0
	}

	return m.interactions[start:]
}

func (m *Memory) GetInsights() map[string]string {
	return m.insights
}

func (m *Memory) GetContext() string {
	if len(m.interactions) == 0 {
		return "Нет истории взаимодействий"
	}

	recent := m.GetRecentInteractions(5)
	context := "История последних действий:\n\n"

	for i, interaction := range recent {
		context += fmt.Sprintf("%d. %s\n", i+1, formatInteraction(interaction))
	}

	if len(m.insights) > 0 {
		context += "\n Insights:\n"
		for key, insight := range m.insights {
			context += fmt.Sprintf("- %s: %s\n", key, insight)
		}
	}

	return context
}

func formatInteraction(interaction Interaction) string {
	return fmt.Sprintf("[%s] %s -> %s",
		interaction.Timestamp.Format("15:04:05"),
		interaction.Action,
		truncateText(interaction.Result, 100))
}

func truncateText(text string, max int) string {
	if len(text) <= max {
		return text
	}
	if max <= 3 {
		return text[:max]
	}
	return text[:max-3] + "..."
}
