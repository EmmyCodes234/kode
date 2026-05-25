package llm

import (
	"fmt"
	"strings"
)

type TokenBudget struct {
	MaxInputTokens     int
	MaxOutputTokens    int
	MaxCostCents       int
}

type modelPricing struct {
	inputCents  int
	outputCents int
}

var modelPrices = map[string]modelPricing{
	"gpt-4o":          {inputCents: 250, outputCents: 1000},
	"gpt-4o-mini":     {inputCents: 15, outputCents: 60},
	"gpt-4":           {inputCents: 3000, outputCents: 6000},
	"gpt-4-turbo":     {inputCents: 1000, outputCents: 3000},
	"gpt-3.5-turbo":   {inputCents: 50, outputCents: 150},
	"claude-3-opus":   {inputCents: 1500, outputCents: 7500},
	"claude-3-sonnet": {inputCents: 300, outputCents: 1500},
	"claude-3-haiku":  {inputCents: 25, outputCents: 125},
}

type BudgetTracker struct {
	model         string
	budget        TokenBudget
	totalInput    int
	totalOutput   int
	exceeded      bool
	exceededMsg   string
}

func NewBudgetTracker(model string, budget TokenBudget) *BudgetTracker {
	return &BudgetTracker{
		model:  model,
		budget: budget,
	}
}

func (bt *BudgetTracker) Track(inputTokens, outputTokens int) {
	bt.totalInput += inputTokens
	bt.totalOutput += outputTokens
	bt.checkExceeded()
}

func (bt *BudgetTracker) checkExceeded() {
	if bt.exceeded {
		return
	}

	if bt.budget.MaxInputTokens > 0 && bt.totalInput > bt.budget.MaxInputTokens {
		bt.exceeded = true
		bt.exceededMsg = fmt.Sprintf("input budget exceeded: %d/%d tokens", bt.totalInput, bt.budget.MaxInputTokens)
		return
	}

	if bt.budget.MaxOutputTokens > 0 && bt.totalOutput > bt.budget.MaxOutputTokens {
		bt.exceeded = true
		bt.exceededMsg = fmt.Sprintf("output budget exceeded: %d/%d tokens", bt.totalOutput, bt.budget.MaxOutputTokens)
		return
	}

	if bt.budget.MaxCostCents > 0 {
		cost := bt.EstimatedCostCents()
		if cost > bt.budget.MaxCostCents {
			bt.exceeded = true
			bt.exceededMsg = fmt.Sprintf("cost budget exceeded: %d¢ / %d¢", cost, bt.budget.MaxCostCents)
			return
		}
	}
}

func (bt *BudgetTracker) EstimatedCostCents() int {
	pricing, ok := modelPrices[bt.model]
	if !ok {
		defaultPrice := modelPricing{inputCents: 250, outputCents: 1000}
		pricing = defaultPrice
	}
	inputCost := (bt.totalInput * pricing.inputCents) / 1000
	outputCost := (bt.totalOutput * pricing.outputCents) / 1000
	return inputCost + outputCost
}

func (bt *BudgetTracker) IsExceeded() bool {
	return bt.exceeded
}

func (bt *BudgetTracker) ExceededMessage() string {
	return bt.exceededMsg
}

func (bt *BudgetTracker) Summary() string {
	cost := bt.EstimatedCostCents()
	parts := []string{
		fmt.Sprintf("in: %d", bt.totalInput),
		fmt.Sprintf("out: %d", bt.totalOutput),
		fmt.Sprintf("cost: %d¢", cost),
	}
	if bt.budget.MaxInputTokens > 0 {
		parts = append(parts, fmt.Sprintf("budget_in: %d", bt.budget.MaxInputTokens))
	}
	if bt.budget.MaxOutputTokens > 0 {
		parts = append(parts, fmt.Sprintf("budget_out: %d", bt.budget.MaxOutputTokens))
	}
	if bt.budget.MaxCostCents > 0 {
		parts = append(parts, fmt.Sprintf("budget_cost: %d¢", bt.budget.MaxCostCents))
	}
	return strings.Join(parts, " | ")
}

func (bt *BudgetTracker) TotalInput() int {
	return bt.totalInput
}

func (bt *BudgetTracker) TotalOutput() int {
	return bt.totalOutput
}
