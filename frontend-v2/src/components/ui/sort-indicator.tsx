import clsx from 'clsx'

export function SortIndicator({
  sorted,
}: {
  sorted: false | 'asc' | 'desc'
}) {
  return (
    <span className={clsx('text-xs', { invisible: !sorted })}>
      {sorted === 'desc' ? '▼' : '▲'}
    </span>
  )
}
