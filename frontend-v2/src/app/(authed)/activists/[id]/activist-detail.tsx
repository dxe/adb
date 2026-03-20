'use client'

import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import {
  API_PATH,
  apiClient,
  ActivistJSON,
  ActivistColumnName,
} from '@/lib/api'
import {
  COLUMN_DEFINITIONS,
  type ColumnDefinition,
} from '../column-definitions'
import { getActivistDisplayName } from '../display-name'
import { formatValue } from '../format-value'
import { LinkedValue } from '../linked-value'

function useActivist(activistId: number) {
  return useQuery({
    queryKey: [API_PATH.ACTIVIST_GET, activistId],
    queryFn: ({ signal }) => apiClient.getActivist(activistId, signal),
  })
}

export function ActivistDetail({ activistId }: { activistId: number }) {
  const { data: activist, isError, isLoading } = useActivist(activistId)

  if (isLoading) {
    return <div className="animate-pulse">Loading activist details...</div>
  }
  if (isError || !activist) {
    return <div>Unable to load activist details.</div>
  }

  const displayName = getActivistDisplayName(activist)

  // Group fields by category using column definitions
  const groupedFields = new Map<
    string,
    { label: string; value: string; linkType?: ColumnDefinition['linkType'] }[]
  >()
  let notes = ''

  for (const def of COLUMN_DEFINITIONS) {
    if (def.hideOnDetailPage) continue

    const rawValue = activist[def.name as keyof ActivistJSON]
    const formatted = formatValue(rawValue, def.name as ActivistColumnName)
    if (!formatted) continue

    if (def.name === 'notes') {
      notes = formatted
      continue
    }

    const group = groupedFields.get(def.category) ?? []
    group.push({ label: def.label, value: formatted, linkType: def.linkType })
    groupedFields.set(def.category, group)
  }

  return (
    <>
      <div className="flex items-center gap-3">
        <Link
          href="/activists"
          className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors"
        >
          {/**
           * Back button does not really go back - just goes to /activists.
           * Detecting if router.back() actually goes back to /activists is
           * complicated but may be implemented in the future to preserve the
           * page state / scroll position.
           */}
          <ArrowLeft className="h-4 w-4" />
          View all Activists
        </Link>
      </div>

      <div className="flex flex-col gap-1">
        <h1
          className={`text-3xl font-bold ${
            displayName.isPlaceholder ? 'italic text-muted-foreground' : ''
          }`}
        >
          {displayName.text}
        </h1>
      </div>

      <div className="flex flex-col gap-8">
        {Array.from(groupedFields.entries()).map(([category, fields]) => (
          <section key={category}>
            <h2 className="text-lg font-semibold mb-3 border-b pb-1">
              {category}
            </h2>
            <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-2">
              {fields.map(({ label, value, linkType }) => (
                <div key={label} className="flex justify-between gap-2 py-1">
                  <dt className="text-sm font-medium text-muted-foreground">
                    {label}
                  </dt>
                  <dd className="text-sm text-right">
                    {linkType ? (
                      <LinkedValue value={value} linkType={linkType} />
                    ) : (
                      value
                    )}
                  </dd>
                </div>
              ))}
            </dl>
          </section>
        ))}

        {/* Notes value may be long, so don't subject it to two-column view. */}
        {notes && (
          <section>
            <h2 className="text-lg font-semibold mb-3 border-b pb-1">Notes</h2>
            <p className="text-sm whitespace-pre-wrap">{notes}</p>
          </section>
        )}
      </div>
    </>
  )
}
