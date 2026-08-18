import { describe, expect, it } from 'vitest'
import type { ActivistJSON } from '@/lib/api'
import { COLUMN_DEFINITION_BY_NAME } from './column-definitions'
import { getReadOnlyFieldDisplay } from './format-value'

describe('getReadOnlyFieldDisplay', () => {
  it('shows assigned_to_name (not the raw id) for the assigned_to field', () => {
    const def = COLUMN_DEFINITION_BY_NAME['assigned_to']
    const activist: ActivistJSON = {
      id: 1,
      assigned_to: 5,
      assigned_to_name: 'Alice Smith',
    }

    const display = getReadOnlyFieldDisplay(activist, def)

    expect(display.value).toBe('Alice Smith')
    expect(display.isEmpty).toBe(false)
  })

  it('shows a placeholder for assigned_to when assigned_to_name is absent', () => {
    const def = COLUMN_DEFINITION_BY_NAME['assigned_to']
    const activist: ActivistJSON = {
      id: 1,
      assigned_to: 5,
    }

    const display = getReadOnlyFieldDisplay(activist, def)

    expect(display.value).toBe('—')
    expect(display.isEmpty).toBe(true)
  })

  it('falls back to formatValue/raw value for a normal (non-assigned_to) field', () => {
    const def = COLUMN_DEFINITION_BY_NAME['name']
    const activist: ActivistJSON = {
      id: 1,
      name: 'Jane Doe',
    }

    const display = getReadOnlyFieldDisplay(activist, def)

    expect(display.value).toBe('Jane Doe')
    expect(display.isEmpty).toBe(false)
  })
})
