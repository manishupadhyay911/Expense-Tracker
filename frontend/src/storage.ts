const FORM_KEY = 'expense-tracker:expense-form'
const FILTER_KEY = 'expense-tracker:expense-filters'
const PENDING_KEY = 'expense-tracker:pending-expense'

export type SavedForm = {
  amount: string
  category: string
  description: string
  date: string
}

export type SavedFilters = {
  category: string
  sort: string
}

export type PendingSubmission = {
  key: string
  payload: SavedForm
}

export function readSavedForm(): SavedForm {
  return readJSON(FORM_KEY, {
    amount: '',
    category: '',
    description: '',
    date: '',
  })
}

export function writeSavedForm(form: SavedForm) {
  writeJSON(FORM_KEY, form)
}

export function clearSavedForm() {
  localStorage.removeItem(FORM_KEY)
}

export function readSavedFilters(): SavedFilters {
  return readJSON(FILTER_KEY, {
    category: '',
    sort: 'date_desc',
  })
}

export function writeSavedFilters(filters: SavedFilters) {
  writeJSON(FILTER_KEY, filters)
}

export function readPendingSubmission(): PendingSubmission | null {
  return readJSON(PENDING_KEY, null)
}

export function writePendingSubmission(pending: PendingSubmission) {
  writeJSON(PENDING_KEY, pending)
}

export function clearPendingSubmission() {
  localStorage.removeItem(PENDING_KEY)
}

function readJSON<T>(key: string, fallback: T): T {
  const raw = localStorage.getItem(key)
  if (!raw) {
    return fallback
  }

  try {
    return JSON.parse(raw) as T
  } catch {
    return fallback
  }
}

function writeJSON(key: string, value: unknown) {
  localStorage.setItem(key, JSON.stringify(value))
}
