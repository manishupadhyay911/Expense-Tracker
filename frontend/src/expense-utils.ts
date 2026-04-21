import type { Expense, ExpenseInput } from './api'

export type CategorySummary = {
  category: string
  totalMinor: number
  count: number
}

export function parseMinorAmount(amount: string): number | null {
  const value = amount.trim()
  if (!/^\d+(\.\d{1,2})?$/.test(value)) {
    return null
  }

  const [wholePart, fractionPart = ''] = value.split('.')
  const whole = Number.parseInt(wholePart, 10)
  const fraction = Number.parseInt((fractionPart + '00').slice(0, 2), 10)

  if (Number.isNaN(whole) || Number.isNaN(fraction)) {
    return null
  }

  return whole * 100 + fraction
}

export function formatRupees(minor: number) {
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(minor / 100)
}

export function totalMinor(expenses: Expense[]) {
  return expenses.reduce((sum, expense) => {
    const parsed = parseMinorAmount(expense.amount)
    return sum + (parsed ?? 0)
  }, 0)
}

export function buildCategorySummaries(expenses: Expense[]): CategorySummary[] {
  const summaries = new Map<string, CategorySummary>()

  for (const expense of expenses) {
    const parsed = parseMinorAmount(expense.amount) ?? 0
    const current = summaries.get(expense.category) ?? {
      category: expense.category,
      totalMinor: 0,
      count: 0,
    }

    current.totalMinor += parsed
    current.count += 1
    summaries.set(expense.category, current)
  }

  return Array.from(summaries.values()).sort((left, right) => left.category.localeCompare(right.category))
}

export function validateExpenseInput(input: ExpenseInput) {
  const amountMinor = parseMinorAmount(input.amount)
  if (amountMinor === null || amountMinor <= 0) {
    return 'Amount must be a positive value with up to two decimal places.'
  }

  if (!input.category.trim()) {
    return 'Category is required.'
  }

  if (!input.description.trim()) {
    return 'Description is required.'
  }

  if (!/^\d{4}-\d{2}-\d{2}$/.test(input.date)) {
    return 'Date is required.'
  }

  return null
}
