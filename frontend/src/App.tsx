import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { createExpense, listExpenses, type Expense, type ExpenseInput } from './api'
import {
  buildCategorySummaries,
  formatRupees,
  parseMinorAmount,
  totalMinor,
  validateExpenseInput,
} from './expense-utils'
import {
  clearPendingSubmission,
  clearSavedForm,
  readPendingSubmission,
  readSavedFilters,
  readSavedForm,
  writePendingSubmission,
  writeSavedFilters,
  writeSavedForm,
} from './storage'

type FormState = ExpenseInput

type FiltersState = {
  category: string
  sort: 'date_desc' | 'date_asc'
}

const DEFAULT_FORM: FormState = {
  amount: '',
  category: '',
  description: '',
  date: '',
}

const DEFAULT_FILTERS: FiltersState = {
  category: '',
  sort: 'date_desc',
}

function buildIdempotencyKey(payload: FormState) {
  return globalThis.crypto?.randomUUID?.() ?? `pending-${Date.now()}-${payload.amount}-${payload.date}`
}

function samePayload(left: FormState, right: FormState) {
  return (
    left.amount === right.amount &&
    left.category === right.category &&
    left.description === right.description &&
    left.date === right.date
  )
}

export default function App() {
  const [form, setForm] = useState<FormState>(() => readSavedForm())
  const [filters, setFilters] = useState<FiltersState>(() => {
    const saved = readSavedFilters()
    return {
      category: saved.category,
      sort: saved.sort === 'date_asc' ? 'date_asc' : 'date_desc',
    }
  })
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [pendingKey, setPendingKey] = useState<string | null>(() => readPendingSubmission()?.key ?? null)
  const [pendingMessage, setPendingMessage] = useState<string | null>(() => {
    const pending = readPendingSubmission()
    return pending ? 'A previous submission is waiting to be retried.' : null
  })

  useEffect(() => {
    let active = true

    async function load() {
      setLoading(true)
      setError(null)
      try {
        const items = await listExpenses(filters.category, filters.sort)
        if (active) {
          setExpenses(items)
        }
      } catch (err) {
        if (active) {
          setError(err instanceof Error ? err.message : 'Failed to load expenses')
        }
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }

    void load()
    return () => {
      active = false
    }
  }, [filters.category, filters.sort])

  useEffect(() => {
    writeSavedForm(form)
  }, [form])

  useEffect(() => {
    writeSavedFilters(filters)
  }, [filters])

  const visibleTotalMinor = useMemo(() => totalMinor(expenses), [expenses])
  const categorySummaries = useMemo(() => buildCategorySummaries(expenses), [expenses])

  const visibleCategories = useMemo(() => {
    const categories = new Set(expenses.map((expense) => expense.category).filter(Boolean))
    return Array.from(categories).sort((left, right) => left.localeCompare(right))
  }, [expenses])

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)
    setSuccess(null)

    const payload = {
      amount: form.amount.trim(),
      category: form.category.trim(),
      description: form.description.trim(),
      date: form.date,
    }

    if (!payload.amount || !payload.category || !payload.description || !payload.date) {
      setError('Please fill amount, category, description, and date.')
      return
    }

    const validationError = validateExpenseInput(payload)
    if (validationError) {
      setError(validationError)
      return
    }

    const pending = readPendingSubmission()
    const key =
      pending && samePayload(pending.payload, payload)
        ? pending.key
        : pendingKey && samePayload(form, payload)
          ? pendingKey
          : buildIdempotencyKey(payload)

    setSubmitting(true)
    setPendingKey(key)
    writePendingSubmission({ key, payload })

    try {
      await createExpense(payload, key)
      clearPendingSubmission()
      clearSavedForm()
      setPendingKey(null)
      setPendingMessage(null)
      setForm(DEFAULT_FORM)
      setSuccess('Expense saved successfully.')
      const items = await listExpenses(filters.category, filters.sort)
      setExpenses(items)
    } catch (err) {
      setPendingMessage('Submission is saved locally. Try again with the same details to retry safely.')
      setError(err instanceof Error ? err.message : 'Failed to create expense')
    } finally {
      setSubmitting(false)
    }
  }

  function updateForm<K extends keyof FormState>(field: K, value: FormState[K]) {
    setForm((current) => ({ ...current, [field]: value }))
  }

  function updateFilters<K extends keyof FiltersState>(field: K, value: FiltersState[K]) {
    setFilters((current) => ({ ...current, [field]: value }))
  }

  return (
    <main className="shell">
      <section className="hero">
        <p className="eyebrow">Expense Tracker</p>
        <h1>Track your spending with a simple, production-ready workflow.</h1>
        <p className="lede">
          Add expenses, filter by category, sort newest-first or oldest-first, and keep an eye on the current
          total. Repeated submits reuse the same idempotency key so retries stay safe.
        </p>
      </section>

      <section className="panel form-panel">
        <div className="panel-header">
          <h2>Add expense</h2>
          <strong>{formatRupees(visibleTotalMinor)}</strong>
        </div>

        <form className="expense-form" onSubmit={handleSubmit}>
          <label>
            <span>Amount</span>
            <input
              type="text"
              inputMode="decimal"
              placeholder="12.50"
              value={form.amount}
              onChange={(event) => updateForm('amount', event.target.value)}
            />
          </label>

          <label>
            <span>Category</span>
            <input
              type="text"
              placeholder="Food"
              value={form.category}
              onChange={(event) => updateForm('category', event.target.value)}
            />
          </label>

          <label className="field-wide">
            <span>Description</span>
            <input
              type="text"
              placeholder="Lunch with client"
              value={form.description}
              onChange={(event) => updateForm('description', event.target.value)}
            />
          </label>

          <label>
            <span>Date</span>
            <input
              type="date"
              value={form.date}
              onChange={(event) => updateForm('date', event.target.value)}
            />
          </label>

          <div className="form-actions">
            <button type="submit" disabled={submitting}>
              {submitting ? 'Saving...' : 'Save expense'}
            </button>
          </div>
        </form>

        {(error || success || pendingMessage) && (
          <div className="status-stack">
            {error && <p className="status status-error">{error}</p>}
            {success && <p className="status status-success">{success}</p>}
            {pendingMessage && <p className="status status-warning">{pendingMessage}</p>}
          </div>
        )}
      </section>

      <section className="panel">
        <div className="panel-header controls-header">
          <div>
            <h2>Current expenses</h2>
            <p className="panel-subtitle">
              {loading
                ? 'Loading expenses...'
                : `${expenses.length} expense${expenses.length === 1 ? '' : 's'} visible`}
            </p>
          </div>

          <div className="controls">
            <label>
              <span>Filter by category</span>
              <select
                value={filters.category}
                onChange={(event) => updateFilters('category', event.target.value)}
              >
                <option value="">All categories</option>
                {visibleCategories.map((category) => (
                  <option key={category} value={category}>
                    {category}
                  </option>
                ))}
              </select>
            </label>

            <label>
              <span>Sort</span>
              <select
                value={filters.sort}
                onChange={(event) => updateFilters('sort', event.target.value as FiltersState['sort'])}
              >
                <option value="date_desc">Date: newest first</option>
                <option value="date_asc">Date: oldest first</option>
              </select>
            </label>
          </div>
        </div>

        <div className="summary-grid">
          {categorySummaries.length === 0 ? (
            <p className="empty-state summary-empty">
              {loading ? 'Loading category totals...' : 'No category totals yet.'}
            </p>
          ) : (
            categorySummaries.map((summary) => (
              <article className="summary-card" key={summary.category}>
                <span>{summary.category}</span>
                <strong>{formatRupees(summary.totalMinor)}</strong>
                <small>{summary.count} expense{summary.count === 1 ? '' : 's'}</small>
              </article>
            ))
          )}
        </div>

        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Date</th>
                <th>Category</th>
                <th>Description</th>
                <th className="numeric">Amount</th>
              </tr>
            </thead>
            <tbody>
              {!loading && expenses.length === 0 && (
                <tr>
                  <td colSpan={4} className="empty-state">
                    No expenses match the current filters.
                  </td>
                </tr>
              )}

              {expenses.map((expense) => (
                <tr key={expense.id}>
                  <td>{expense.date}</td>
                  <td>{expense.category}</td>
                  <td>{expense.description}</td>
                  <td className="numeric">{formatRupees(parseMinorAmount(expense.amount) ?? 0)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </main>
  )
}
