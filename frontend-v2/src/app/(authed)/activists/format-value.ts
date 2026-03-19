import { z } from 'zod'
import { ActivistJSON, ActivistColumnName } from '@/lib/api'
import { COLUMN_DEFINITIONS } from './column-definitions'
import { formatDateValueForActivists } from './date-time'

// Gets the underlying type of a column from the ActivistJSON schema
const getColumnType = (
  columnName: ActivistColumnName,
): 'string' | 'number' | 'boolean' => {
  const schema = ActivistJSON.shape[
    columnName as keyof typeof ActivistJSON.shape
  ] as z.ZodTypeAny
  if (!schema) throw new Error('column not in schema: ' + columnName)

  const unwrapped =
    schema instanceof z.ZodOptional || schema instanceof z.ZodNullable
      ? schema.unwrap()
      : schema

  if (unwrapped instanceof z.ZodNumber) return 'number'
  if (unwrapped instanceof z.ZodBoolean) return 'boolean'
  return 'string'
}

export const formatValue = (
  value: unknown,
  columnName: ActivistColumnName,
): string => {
  if (value === null || value === undefined) return ''

  const columnType = getColumnType(columnName)

  if (columnType === 'boolean') {
    return value ? 'Yes' : 'No'
  }

  if (columnType === 'number') {
    return String(value)
  }

  if (columnType === 'string') {
    const definition = COLUMN_DEFINITIONS.find((d) => d.name === columnName)
    if (definition?.isDate && typeof value === 'string') {
      if (
        /^\d{4}-\d{2}-\d{2}$/.test(value) ||
        !isNaN(new Date(value).getTime())
      ) {
        return formatDateValueForActivists(value)
      }
    }
  }

  return String(value)
}
