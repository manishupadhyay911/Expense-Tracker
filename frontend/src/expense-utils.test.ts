import { describe, expect, it } from 'vitest'
import type { Expense } from './api'
import {
  buildCategorySummaries,
  parseMinorAmount,
  totalMinor,
  validateExpenseInput,
} from './expense-utils'

describe('expense-utils', () => {
  it('parses valid money amounts and rejects negative values', () => {
    expect(parseMinorAmount('12.50')).toBe(1250)
    expect(parseMinorAmount('-12.50')).toBeNull()
    expect(parseMinorAmount('12.505')).toBeNull()
  })

  it('builds category summaries and totals', () => {
    const expenses: Expense[] = [
      {
        id: 1,
        amount: '10.00',
        category: 'Food',
        description: 'Lunch',
        date: '2026-04-21',
        created_at: '2026-04-21T10:00:00Z',
      },
      {
        id: 2,
        amount: '15.50',
        category: 'Travel',
        description: 'Cab',
        date: '2026-04-21',
        created_at: '2026-04-21T11:00:00Z',
      },
      {
        id: 3,
        amount: '5.25',
        category: 'Food',
        description: 'Tea',
        date: '2026-04-20',
        created_at: '2026-04-20T11:00:00Z',
      },
    ]

    expect(totalMinor(expenses)).toBe(3075)
    expect(buildCategorySummaries(expenses)).toEqual([
      { category: 'Food', totalMinor: 1525, count: 2 },
      { category: 'Travel', totalMinor: 1550, count: 1 },
    ])
  })

  it('validates expense input', () => {
    expect(
      validateExpenseInput({
        amount: '-10',
        category: 'Food',
        description: 'Lunch',
        date: '2026-04-21',
      }),
    ).toContain('positive value')

    expect(
      validateExpenseInput({
        amount: '10.00',
        category: '',
        description: 'Lunch',
        date: '2026-04-21',
      }),
    ).toBe('Category is required.')
  })
})
