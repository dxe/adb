import { z } from 'zod'
import { ActivistJSON, ActivistColumnName } from '@/lib/api'
import { COLUMN_DEFINITION_BY_NAME } from './column-definitions'
import { formatDateValueForActivists } from './date-time'

type ColumnType = 'string' | 'number' | 'boolean'

const COLUMN_TYPE_BY_NAME = Object.fromEntries(
  Object.entries(ActivistJSON.shape).map(([columnName, schema]) => {
    const unwrapped =
      schema instanceof z.ZodOptional || schema instanceof z.ZodNullable
        ? schema.unwrap()
        : schema

    let columnType: ColumnType = 'string'
    if (unwrapped instanceof z.ZodNumber) {
      columnType = 'number'
    } else if (unwrapped instanceof z.ZodBoolean) {
      columnType = 'boolean'
    }

    return [columnName, columnType]
  }),
) as Record<ActivistColumnName, ColumnType>

export const formatValue = (
  value: unknown,
  columnName: ActivistColumnName,
): string => {
  if (value === null || value === undefined) return ''

  const columnType = COLUMN_TYPE_BY_NAME[columnName]

  if (columnType === 'boolean') {
    return value ? 'Yes' : 'No'
  }

  if (columnType === 'number') {
    return String(value)
  }

  if (columnType === 'string') {
    const definition = COLUMN_DEFINITION_BY_NAME[columnName]
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
