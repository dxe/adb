'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { Pencil } from 'lucide-react'
import { ActivistJSON } from '@/lib/api'
import { Button } from '@/components/ui/button'
import {
  COLUMN_DEFINITIONS,
  isEditableActivistField,
  type ColumnCategory,
  type ColumnDefinition,
} from '../column-definitions'
import { FieldDescriptionPopover } from '../field-description-popover'
import {
  formatValue,
  getReadOnlyFieldDisplay,
  type ReadOnlyFieldDisplay,
} from '../format-value'
import { LinkedValue } from '../linked-value'
import { ActivistSectionForm } from './section-form'

const NOTES_SECTION_KEY = '__notes__'
type SectionKey = ColumnCategory | typeof NOTES_SECTION_KEY

// All fields displayed for each category. Includes both editable and
// non-editable fields so the edit-mode grid keeps the same shape as the
// read-only grid and fields don't jump around when Edit is clicked.
const SECTION_FIELDS_BY_CATEGORY: Map<ColumnCategory, ColumnDefinition[]> =
  (() => {
    const map = new Map<ColumnCategory, ColumnDefinition[]>()
    for (const def of COLUMN_DEFINITIONS) {
      if (def.hideOnDetailPage) continue
      // Notes is rendered as its own section, not as part of "Other".
      if (def.name === 'notes') continue
      const list = map.get(def.category) ?? []
      list.push(def)
      map.set(def.category, list)
    }
    return map
  })()

const EDITABLE_CATEGORIES: Set<ColumnCategory> = new Set(
  Array.from(SECTION_FIELDS_BY_CATEGORY.entries())
    .filter(([, defs]) => defs.some((d) => isEditableActivistField(d.name)))
    .map(([cat]) => cat),
)

const NOTES_DEFINITION = COLUMN_DEFINITIONS.find((d) => d.name === 'notes')!
if (!NOTES_DEFINITION) {
  throw new Error("Column definition for 'notes' is missing")
}

const SECTION_ORDER: ColumnCategory[] = (() => {
  const order: ColumnCategory[] = []
  const seen = new Set<ColumnCategory>()
  for (const def of COLUMN_DEFINITIONS) {
    if (seen.has(def.category)) continue
    seen.add(def.category)
    order.push(def.category)
  }
  return order
})()

export function ActivistDetails({
  activistId,
  activist,
  onDirtyChange,
}: {
  activistId: number
  activist: ActivistJSON
  onDirtyChange: (isDirty: boolean) => void
}) {
  const [editingSection, setEditingSection] = useState<SectionKey | null>(null)
  const [isFormDirty, setIsFormDirty] = useState(false)

  const setDirty = useCallback(
    (isDirty: boolean) => {
      setIsFormDirty(isDirty)
      onDirtyChange(isDirty)
    },
    [onDirtyChange],
  )

  const confirmDiscard = useCallback(() => {
    if (!isFormDirty) return true
    return window.confirm(
      'You have unsaved changes. Discard them and leave this section?',
    )
  }, [isFormDirty])

  const handleEdit = useCallback(
    (section: SectionKey) => {
      if (editingSection !== null && editingSection !== section) {
        if (!confirmDiscard()) return
      }
      setEditingSection(section)
      setDirty(false)
    },
    [editingSection, confirmDiscard, setDirty],
  )

  const handleCancel = useCallback(() => {
    setEditingSection(null)
    setDirty(false)
  }, [setDirty])

  const handleSaved = useCallback(() => {
    setEditingSection(null)
    setDirty(false)
  }, [setDirty])

  // Warn on full page unload (close/refresh) while edits are unsaved.
  useEffect(() => {
    if (!isFormDirty) return
    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault()
      // Required for older browsers that read returnValue.
      e.returnValue = ''
    }
    window.addEventListener('beforeunload', handler)
    return () => window.removeEventListener('beforeunload', handler)
  }, [isFormDirty])

  const groupedFields = useMemo(() => buildReadOnlyFields(activist), [activist])

  const notesValue = useMemo(
    () => formatValue(activist.notes, 'notes'),
    [activist],
  )

  return (
    <div className="flex flex-col gap-8">
      {SECTION_ORDER.map((category) => {
        const fields = groupedFields.get(category) ?? []
        const sectionFields = SECTION_FIELDS_BY_CATEGORY.get(category)
        const canEdit = EDITABLE_CATEGORIES.has(category)
        if (fields.length === 0 && !sectionFields) return null
        const isEditing = editingSection === category
        return (
          <section key={category}>
            <SectionHeader
              title={category}
              showEdit={canEdit && editingSection === null}
              onEdit={() => handleEdit(category)}
            />
            {isEditing && sectionFields ? (
              <ActivistSectionForm
                activistId={activistId}
                activist={activist}
                fields={sectionFields}
                onSaved={handleSaved}
                onCancel={handleCancel}
                onDirtyChange={setDirty}
              />
            ) : (
              <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-2">
                {fields.map(
                  ({ label, value, description, linkType, isEmpty }) => (
                    <div
                      key={label}
                      className="flex justify-between gap-2 py-1"
                    >
                      <dt
                        className={`flex items-center gap-1 text-sm font-medium text-muted-foreground ${
                          isEmpty ? 'opacity-50' : ''
                        }`}
                      >
                        {label}
                        {description && (
                          <FieldDescriptionPopover
                            label={label}
                            description={description}
                          />
                        )}
                      </dt>
                      <dd
                        className={`text-sm text-right ${
                          isEmpty ? 'text-muted-foreground opacity-50' : ''
                        }`}
                      >
                        {!isEmpty && linkType ? (
                          <LinkedValue value={value} linkType={linkType} />
                        ) : (
                          value
                        )}
                      </dd>
                    </div>
                  ),
                )}
              </dl>
            )}
          </section>
        )
      })}

      {/* Notes is its own section so its (potentially long) value can use the
          full width in both read and edit modes. */}
      <section>
        <SectionHeader
          title="Notes"
          showEdit={editingSection === null}
          onEdit={() => handleEdit(NOTES_SECTION_KEY)}
        />
        {editingSection === NOTES_SECTION_KEY ? (
          <ActivistSectionForm
            activistId={activistId}
            activist={activist}
            fields={[NOTES_DEFINITION]}
            onSaved={handleSaved}
            onCancel={handleCancel}
            onDirtyChange={setDirty}
          />
        ) : notesValue ? (
          <p className="text-sm whitespace-pre-wrap">{notesValue}</p>
        ) : (
          <p className="text-sm text-muted-foreground italic">No notes</p>
        )}
      </section>
    </div>
  )
}

function buildReadOnlyFields(
  activist: ActivistJSON,
): Map<ColumnCategory, ReadOnlyFieldDisplay[]> {
  const grouped = new Map<ColumnCategory, ReadOnlyFieldDisplay[]>()
  for (const def of COLUMN_DEFINITIONS) {
    if (def.hideOnDetailPage) continue
    if (def.name === 'notes') continue
    const group = grouped.get(def.category) ?? []
    group.push(getReadOnlyFieldDisplay(activist, def))
    grouped.set(def.category, group)
  }
  return grouped
}

function SectionHeader({
  title,
  showEdit,
  onEdit,
}: {
  title: string
  showEdit: boolean
  onEdit: () => void
}) {
  return (
    <div className="mb-3 flex items-center justify-between border-b pb-1">
      <h2 className="text-lg font-semibold">{title}</h2>
      {showEdit && (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onEdit}
          className="h-7 px-2"
          aria-label={`Edit ${title}`}
        >
          <Pencil className="h-3.5 w-3.5" />
          Edit
        </Button>
      )}
    </div>
  )
}
