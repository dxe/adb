'use client'

import { ReactNode } from 'react'
import { Checkbox } from '@/components/ui/checkbox'
import { cn } from '@/lib/utils'

/** A statement followed by a bordered "Yes, I agree" checkbox that turns green when checked. */
export function AgreementCheckbox({
  id,
  checked,
  onCheckedChange,
  error,
  label = 'Yes, I agree.',
  children,
}: {
  id: string
  checked: boolean
  onCheckedChange: (checked: boolean) => void
  error?: string
  label?: string
  children: ReactNode
}) {
  return (
    <div className="flex flex-col gap-2">
      <p id={`${id}-statement`} className="text-sm">
        {children}
      </p>
      <label
        htmlFor={id}
        className={cn(
          'flex w-fit cursor-pointer items-center gap-2 rounded-md border p-2 text-sm transition-colors',
          checked ? 'border-green-600 bg-green-50' : 'hover:bg-accent',
        )}
      >
        <Checkbox
          id={id}
          aria-describedby={`${id}-statement`}
          checked={checked}
          onCheckedChange={(value) => onCheckedChange(Boolean(value))}
          className="data-[state=checked]:border-green-600 data-[state=checked]:bg-green-600 data-[state=checked]:text-white"
        />
        {label}
      </label>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}
