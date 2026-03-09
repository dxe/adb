import clsx from 'clsx'

export function SortIndicator({ sorted }: { sorted: false | 'asc' | 'desc' }) {
  return (
    <span className={clsx('text-xs', { invisible: !sorted })}>
      <span aria-hidden="true">{sorted === 'desc' ? '▼' : '▲'}</span>
      {sorted && (
        <span className="sr-only">
          {sorted === 'desc' ? 'sorted descending' : 'sorted ascending'}
        </span>
      )}
    </span>
  )
}
