'use client'

interface IncludeExcludeOptionProps {
  label: string
  selected: boolean
  mode: 'include' | 'exclude'
  onToggle: () => void
}

/**
 * A toggleable option in an include/exclude filter. When selected, its box
 * shows "+" in the primary color for include, or "-" in red for exclude.
 */
export function IncludeExcludeOption({
  label,
  selected,
  mode,
  onToggle,
}: IncludeExcludeOptionProps) {
  return (
    <button
      type="button"
      role="checkbox"
      aria-checked={selected}
      className="flex w-full items-center gap-2 rounded px-2 py-1 text-left text-sm hover:bg-muted transition-colors"
      onClick={onToggle}
    >
      <span
        aria-hidden
        className={`flex h-4 w-4 shrink-0 items-center justify-center rounded border text-xs font-bold ${
          selected
            ? mode === 'include'
              ? 'bg-primary text-primary-foreground border-primary'
              : 'bg-destructive text-destructive-foreground border-destructive'
            : 'border-input'
        }`}
      >
        {selected ? (mode === 'include' ? '+' : '-') : ''}
      </span>
      {label}
    </button>
  )
}
