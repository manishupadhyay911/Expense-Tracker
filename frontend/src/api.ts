export type Expense = {
  id: number
  amount: string
  category: string
  description: string
  date: string
  created_at: string
}

export type ExpenseInput = {
  amount: string
  category: string
  description: string
  date: string
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api'

function buildUrl(path: string) {
  return `${API_BASE_URL}${path}`
}

export async function listExpenses(category: string, sort: string): Promise<Expense[]> {
  const params = new URLSearchParams()
  if (category.trim()) {
    params.set('category', category.trim())
  }
  if (sort.trim()) {
    params.set('sort', sort.trim())
  }

  const query = params.toString()
  const response = await fetch(buildUrl(`/expenses${query ? `?${query}` : ''}`))
  if (!response.ok) {
    throw new Error(`Failed to load expenses (${response.status})`)
  }

  return response.json()
}

export async function createExpense(
  input: ExpenseInput,
  idempotencyKey: string,
): Promise<Expense> {
  const response = await fetch(buildUrl('/expenses'), {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Idempotency-Key': idempotencyKey,
    },
    body: JSON.stringify(input),
  })

  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as { error?: string } | null
    throw new Error(payload?.error ?? `Failed to create expense (${response.status})`)
  }

  return response.json()
}

