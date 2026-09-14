'use client'

import { ActivistJSON, ActivistColumnName } from '@/lib/api'
import { Check, Minus } from 'lucide-react'
import { IntentPrefetchLink } from '@/components/intent-prefetch-link'
import { useLongPress } from '@/hooks/use-long-press'
import { cn } from '@/lib/utils'
import { COLUMN_DEFINITION_BY_NAME } from './column-definitions'
import { getActivistDisplayName } from './display-name'
import { formatValue, COLUMN_TYPE_BY_NAME } from './format-value'
import type { ActivistSelection } from './use-activist-selection'

interface ActivistCardProps {
  activist: ActivistJSON
  visibleColumns: ActivistColumnName[]
  onActivistClick?: (id: number) => void
  isStale: boolean
  selection?: ActivistSelection
}

/**
 * One activist as a card, for the mobile layout. Long-pressing the card
 * selects it; while anything is selected, tapping toggles selection instead of
 * opening the activist.
 */
export function ActivistCard({
  activist,
  visibleColumns,
  onActivistClick,
  isStale,
  selection,
}: ActivistCardProps) {
  const isSelected = selection?.selectedIds.has(activist.id) ?? false
  const isSelectionMode = (selection?.selectedIds.size ?? 0) > 0

  const { handlers, consumeLongPress } = useLongPress(() =>
    selection?.onToggle(activist.id),
  )

  const displayName = getActivistDisplayName(activist)
  const cardClass = cn(
    'block w-full rounded-lg border bg-card p-4 text-left transition-colors hover:border-primary/50',
    isStale && 'opacity-60',
    // Long-pressing a link otherwise starts a text selection and pops the
    // platform callout menu on top of the gesture.
    selection && 'select-none [-webkit-touch-callout:none]',
    isSelected && 'border-primary bg-primary/15 ring-1 ring-primary/40',
  )

  const cardContent = (
    <div className="flex flex-col gap-2">
      {visibleColumns.map((colName) => {
        const definition = COLUMN_DEFINITION_BY_NAME[colName]
        const label = definition?.label || colName
        const isBool = COLUMN_TYPE_BY_NAME[colName] === 'boolean'
        const rawValue = activist[colName as keyof ActivistJSON]
        const formattedValue = isBool
          ? null
          : colName === 'name'
            ? displayName.text
            : formatValue(rawValue, colName)

        return (
          <div key={colName} className="flex justify-between gap-2">
            <span className="text-sm font-medium text-muted-foreground">
              {label}:
            </span>
            {isBool ? (
              rawValue ? (
                <Check className="h-4 w-4 text-foreground" />
              ) : (
                <Minus className="h-4 w-4 text-muted-foreground" />
              )
            ) : (
              <span
                className={`text-sm ${
                  colName === 'name' && displayName.isPlaceholder
                    ? 'italic text-muted-foreground'
                    : ''
                }`}
              >
                {formattedValue}
              </span>
            )}
          </div>
        )
      })}
    </div>
  )

  const selectionProps = selection
    ? {
        ...handlers,
        'aria-pressed': isSelectionMode ? isSelected : undefined,
      }
    : {}

  // Returns true if the click was consumed by selection and must not navigate.
  const handleSelectionClick = () => {
    if (!selection) return false
    // The long press already toggled this card; swallow its trailing click.
    if (consumeLongPress()) return true
    if (isSelectionMode) {
      selection.onToggle(activist.id)
      return true
    }
    return false
  }

  return onActivistClick ? (
    <a
      data-testid={`activist-card-${activist.id}`}
      href={`/v2/activists/${activist.id}`}
      className={cardClass}
      {...selectionProps}
      onClick={(e) => {
        if (handleSelectionClick()) {
          e.preventDefault()
          return
        }
        if (e.ctrlKey || e.metaKey || e.shiftKey) return
        e.preventDefault()
        onActivistClick(activist.id)
      }}
    >
      {cardContent}
    </a>
  ) : (
    <IntentPrefetchLink
      data-testid={`activist-card-${activist.id}`}
      href={`/activists/${activist.id}`}
      className={cardClass}
      {...selectionProps}
      onClick={(e) => {
        if (handleSelectionClick()) e.preventDefault()
      }}
    >
      {cardContent}
    </IntentPrefetchLink>
  )
}
