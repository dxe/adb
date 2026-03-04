/**
 * Parse a base-10 integer token safely.
 * Returns undefined for non-integer tokens or values outside JS safe integer range.
 */
export function parseSafeInteger(value: string): number | undefined {
  if (!/^-?\d+$/.test(value)) return undefined
  const n = Number(value)
  return Number.isSafeInteger(n) ? n : undefined
}

export function parseOptionalSafeInteger(value?: string): number | undefined {
  if (value === undefined) return value
  return parseSafeInteger(value)
}
