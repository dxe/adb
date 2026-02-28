/** Parse "a,b,-c" into include/exclude sets. */
export function parseIncludeExclude(value?: string): {
  include: Set<string>
  exclude: Set<string>
} {
  const include = new Set<string>()
  const exclude = new Set<string>()
  if (!value) return { include, exclude }
  for (const part of value.split(',')) {
    const trimmed = part.trim()
    if (!trimmed) continue
    if (trimmed.startsWith('-')) {
      exclude.add(trimmed.slice(1))
    } else {
      include.add(trimmed)
    }
  }
  return { include, exclude }
}

/** Build "a,b,-c" from include/exclude sets. */
export function buildIncludeExclude(
  include: Set<string>,
  exclude: Set<string>,
): string | undefined {
  const parts = [
    ...Array.from(include),
    ...Array.from(exclude).map((v) => `-${v}`),
  ]
  return parts.length > 0 ? parts.join(',') : undefined
}

/** Parse "=a,b" or "=~a,b" into mode + values. */
export function parseActivistLevelValue(value?: string): {
  mode: 'include' | 'exclude'
  values: Set<string>
} {
  const empty = { mode: 'include' as const, values: new Set<string>() }
  if (!value) return empty

  if (value.startsWith('=')) {
    const payload = value.slice(1)
    const mode: 'include' | 'exclude' = payload.startsWith('~')
      ? 'exclude'
      : 'include'
    const rawValues = mode === 'exclude' ? payload.slice(1) : payload
    const values = new Set(
      rawValues
        .split(',')
        .map((v) => v.trim())
        .filter(Boolean),
    )
    return { mode, values }
  }

  // Backward compatibility for legacy URL values.
  const { include, exclude } = parseIncludeExclude(value)
  if (exclude.size > 0 && include.size === 0) {
    return { mode: 'exclude', values: exclude }
  }
  if (include.size > 0) {
    return { mode: 'include', values: include }
  }
  return empty
}

/** Build "=a,b" or "=~a,b" from mode + values. */
export function buildActivistLevelValue(
  mode: 'include' | 'exclude',
  values: Set<string>,
): string | undefined {
  const list = Array.from(values)
  if (list.length === 0) return undefined
  return mode === 'exclude' ? `=~${list.join(',')}` : `=${list.join(',')}`
}

/** Parse "1..4" into parts. */
export function parseIntRange(value?: string): { gte?: string; lt?: string } {
  if (!value) return {}
  const parts = value.split('..')
  if (parts.length !== 2) return {}
  return { gte: parts[0] || undefined, lt: parts[1] || undefined }
}

/** Build "1..4" from parts. */
export function buildIntRange(gte?: string, lt?: string): string | undefined {
  if (!gte && !lt) return undefined
  return `${gte || ''}..${lt || ''}`
}
