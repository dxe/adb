'use client'

import { useEffect, useMemo } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { Loader2, Save, X } from 'lucide-react'
import {
  API_PATH,
  apiClient,
  ActivistJSON,
  ActivistPatchInput,
} from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Checkbox } from '@/components/ui/checkbox'
import { DatePicker } from '@/components/ui/date-picker'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  isEditableActivistField,
  type ActivistEditInputType,
  type ColumnDefinition,
} from '../column-definitions'
import { datePickerValueToYmd, ymdToDatePickerValue } from '../date-time'
import { FieldDescriptionPopover } from '../field-description-popover'
import { formatValue } from '../format-value'
import { LinkedValue } from '../linked-value'

type FieldValue = string | boolean | number
type FormValues = Record<string, FieldValue>

// assigned_to uses 0 to mean "unassigned" — the column is non-nullable.
const UNASSIGNED_USER_ID = 0

// Radix Select disallows empty SelectItem values, so we use a sentinel for
// enum-select fields whose editOptions include the empty string.
const ENUM_EMPTY_SENTINEL = '__empty__'

function LabelRow({
  htmlFor,
  label,
  description,
}: {
  htmlFor: string
  label: string
  description?: string
}) {
  return (
    <div className="flex items-center gap-1">
      <Label htmlFor={htmlFor}>{label}</Label>
      {description && (
        <FieldDescriptionPopover label={label} description={description} />
      )}
    </div>
  )
}

const inputTypeFor = (def: ColumnDefinition): ActivistEditInputType =>
  def.editInputType ?? 'text'

const initialValueFor = (
  activist: ActivistJSON,
  def: ColumnDefinition,
): FieldValue => {
  const raw = (activist as Record<string, unknown>)[def.name]
  switch (inputTypeFor(def)) {
    case 'checkbox':
      return Boolean(raw)
    case 'user-select':
      return typeof raw === 'number' ? raw : UNASSIGNED_USER_ID
    case 'date':
      // <input type="date"> needs YYYY-MM-DD; the API returns either that
      // or a longer ISO timestamp.
      return typeof raw === 'string' && raw.length >= 10 ? raw.slice(0, 10) : ''
    default:
      return typeof raw === 'string' ? raw : ''
  }
}

interface ActivistSectionFormProps {
  activistId: number
  activist: ActivistJSON
  fields: ColumnDefinition[]
  onSaved: () => void
  onCancel: () => void
  onDirtyChange: (dirty: boolean) => void
}

export function ActivistSectionForm({
  activistId,
  activist,
  fields,
  onSaved,
  onCancel,
  onDirtyChange,
}: ActivistSectionFormProps) {
  const queryClient = useQueryClient()

  const editableFields = useMemo(
    () => fields.filter((f) => isEditableActivistField(f.name)),
    [fields],
  )

  const hasUserSelect = editableFields.some(
    (f) => inputTypeFor(f) === 'user-select',
  )
  const usersQuery = useQuery({
    queryKey: [API_PATH.USERS],
    queryFn: ({ signal }) => apiClient.getUsers(signal),
    enabled: hasUserSelect,
  })

  const initialValues = useMemo<FormValues>(() => {
    const obj: FormValues = {}
    for (const f of editableFields) obj[f.name] = initialValueFor(activist, f)
    return obj
  }, [activist, editableFields])

  const mutation = useMutation({
    mutationFn: (patch: ActivistPatchInput) =>
      apiClient.patchActivist(activistId, patch),
    onSuccess: (updated) => {
      queryClient.setQueryData([API_PATH.ACTIVIST_GET, activistId], updated)
      queryClient.invalidateQueries({ queryKey: [API_PATH.ACTIVISTS_SEARCH] })
      toast.success('Saved')
      onSaved()
    },
    onError: (err: unknown) => {
      const msg = err instanceof Error ? err.message : 'Failed to save'
      toast.error(msg)
    },
  })

  const form = useForm({
    defaultValues: initialValues,
    onSubmit: async ({ value }) => {
      const patch: Record<string, unknown> = {}
      for (const f of editableFields) {
        if (
          form.state.fieldMeta[f.name]?.isDirty &&
          !Object.is(initialValues[f.name], value[f.name])
        ) {
          patch[f.name] = value[f.name]
        }
      }
      if (Object.keys(patch).length === 0) {
        onSaved()
        return
      }
      await mutation.mutateAsync(patch as ActivistPatchInput)
    },
  })

  const isDirty = useStore(form.store, (state) => state.isDirty)
  useEffect(() => {
    onDirtyChange(isDirty)
    return () => onDirtyChange(false)
  }, [isDirty, onDirtyChange])

  const isSaving = mutation.isPending

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        form.handleSubmit()
      }}
      className="flex flex-col gap-4"
    >
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-4">
        {fields.map((def) => {
          const inputId = `activist-field-${def.name}`
          const inputType = inputTypeFor(def)
          if (!isEditableActivistField(def.name)) {
            return (
              <ReadOnlyField key={def.name} def={def} activist={activist} />
            )
          }
          return (
            <form.Field key={def.name} name={def.name}>
              {(field) => {
                const error = field.state.meta.errors[0]
                const errorMessage =
                  typeof error === 'string'
                    ? error
                    : (error as { message?: string } | undefined)?.message

                if (inputType === 'checkbox') {
                  return (
                    <div className="space-y-1 sm:col-span-1">
                      <div className="flex items-center gap-1">
                        <Label
                          htmlFor={inputId}
                          className="flex items-center gap-2 text-sm"
                        >
                          <Checkbox
                            id={inputId}
                            checked={Boolean(field.state.value)}
                            onCheckedChange={(checked) =>
                              field.handleChange(Boolean(checked))
                            }
                            disabled={isSaving}
                          />
                          {def.label}
                        </Label>
                        {def.description && (
                          <FieldDescriptionPopover
                            label={def.label}
                            description={def.description}
                          />
                        )}
                      </div>
                      {errorMessage && (
                        <p className="text-sm text-destructive">
                          {errorMessage}
                        </p>
                      )}
                    </div>
                  )
                }

                if (inputType === 'enum-select') {
                  const stringValue =
                    typeof field.state.value === 'string'
                      ? field.state.value
                      : ''
                  return (
                    <div className="space-y-1">
                      <LabelRow
                        htmlFor={inputId}
                        label={def.label}
                        description={def.description}
                      />
                      <Select
                        value={
                          stringValue === '' ? ENUM_EMPTY_SENTINEL : stringValue
                        }
                        onValueChange={(value) =>
                          field.handleChange(
                            value === ENUM_EMPTY_SENTINEL ? '' : value,
                          )
                        }
                        disabled={isSaving}
                      >
                        <SelectTrigger id={inputId}>
                          <SelectValue
                            placeholder={`Select ${def.label.toLowerCase()}`}
                          />
                        </SelectTrigger>
                        <SelectContent>
                          {def.editOptions?.map((option) => (
                            <SelectItem
                              key={option}
                              value={
                                option === '' ? ENUM_EMPTY_SENTINEL : option
                              }
                            >
                              {option === '' ? '—' : option}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      {errorMessage && (
                        <p className="text-sm text-destructive">
                          {errorMessage}
                        </p>
                      )}
                    </div>
                  )
                }

                if (inputType === 'user-select') {
                  const numericValue =
                    typeof field.state.value === 'number'
                      ? field.state.value
                      : UNASSIGNED_USER_ID
                  return (
                    <div className="space-y-1">
                      <LabelRow
                        htmlFor={inputId}
                        label={def.label}
                        description={def.description}
                      />
                      <Select
                        value={String(numericValue)}
                        onValueChange={(value) => {
                          const next = parseInt(value, 10)
                          field.handleChange(
                            Number.isNaN(next) ? UNASSIGNED_USER_ID : next,
                          )
                        }}
                        disabled={
                          isSaving || usersQuery.isLoading || usersQuery.isError
                        }
                      >
                        <SelectTrigger id={inputId}>
                          <SelectValue placeholder="Select a user" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value={String(UNASSIGNED_USER_ID)}>
                            Unassigned
                          </SelectItem>
                          {usersQuery.data?.map((user) => (
                            <SelectItem key={user.id} value={String(user.id)}>
                              {user.name || user.email}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      {usersQuery.isError && (
                        <p className="text-sm text-destructive">
                          Failed to load users
                        </p>
                      )}
                      {errorMessage && (
                        <p className="text-sm text-destructive">
                          {errorMessage}
                        </p>
                      )}
                    </div>
                  )
                }

                if (inputType === 'date') {
                  const stringValue =
                    typeof field.state.value === 'string'
                      ? field.state.value
                      : ''
                  return (
                    <div className="space-y-1">
                      <LabelRow
                        htmlFor={inputId}
                        label={def.label}
                        description={def.description}
                      />
                      <DatePicker
                        value={
                          stringValue
                            ? ymdToDatePickerValue(stringValue)
                            : undefined
                        }
                        onValueChange={(date) =>
                          field.handleChange(
                            date ? (datePickerValueToYmd(date) ?? '') : '',
                          )
                        }
                        disabled={isSaving}
                      />
                      {errorMessage && (
                        <p className="text-sm text-destructive">
                          {errorMessage}
                        </p>
                      )}
                    </div>
                  )
                }

                if (inputType === 'textarea') {
                  return (
                    <div className="space-y-1 sm:col-span-2">
                      <LabelRow
                        htmlFor={inputId}
                        label={def.label}
                        description={def.description}
                      />
                      <Textarea
                        id={inputId}
                        rows={6}
                        value={String(field.state.value ?? '')}
                        onChange={(e) => field.handleChange(e.target.value)}
                        onBlur={field.handleBlur}
                        disabled={isSaving}
                      />
                      {errorMessage && (
                        <p className="text-sm text-destructive">
                          {errorMessage}
                        </p>
                      )}
                    </div>
                  )
                }

                const htmlInputType =
                  def.linkType === 'mailto'
                    ? 'email'
                    : def.linkType === 'tel'
                      ? 'tel'
                      : def.linkType === 'url'
                        ? 'url'
                        : 'text'

                return (
                  <div className="space-y-1">
                    <LabelRow
                      htmlFor={inputId}
                      label={def.label}
                      description={def.description}
                    />
                    <Input
                      id={inputId}
                      type={htmlInputType}
                      value={String(field.state.value ?? '')}
                      onChange={(e) => field.handleChange(e.target.value)}
                      onBlur={field.handleBlur}
                      disabled={isSaving}
                    />
                    {errorMessage && (
                      <p className="text-sm text-destructive">{errorMessage}</p>
                    )}
                  </div>
                )
              }}
            </form.Field>
          )
        })}
      </div>

      <div className="flex items-center justify-end gap-2">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={isSaving}
        >
          <X className="h-4 w-4" />
          Cancel
        </Button>
        <Button type="submit" disabled={isSaving}>
          {isSaving ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" />
              Saving...
            </>
          ) : (
            <>
              <Save className="h-4 w-4" />
              Save
            </>
          )}
        </Button>
      </div>
    </form>
  )
}

function ReadOnlyField({
  def,
  activist,
}: {
  def: ColumnDefinition
  activist: ActivistJSON
}) {
  const raw = (activist as Record<string, unknown>)[def.name]
  const formatted = formatValue(raw, def.name)
  const isEmpty = !raw
  return (
    <div className="space-y-1">
      <div className="flex items-center gap-1">
        <span className="text-sm font-medium leading-none">{def.label}</span>
        {def.description && (
          <FieldDescriptionPopover
            label={def.label}
            description={def.description}
          />
        )}
      </div>
      <div
        className={`flex h-9 items-center text-sm ${
          isEmpty ? 'text-muted-foreground opacity-50' : 'text-muted-foreground'
        }`}
      >
        {isEmpty ? (
          '—'
        ) : def.linkType ? (
          <LinkedValue value={formatted} linkType={def.linkType} />
        ) : (
          formatted
        )}
      </div>
    </div>
  )
}
