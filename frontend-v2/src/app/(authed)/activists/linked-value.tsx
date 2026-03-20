import type { ColumnDefinition } from './column-definitions'

type LinkType = NonNullable<ColumnDefinition['linkType']>

function getLinkHref(value: string, linkType: LinkType): string | null {
  if (value.trim() === '') {
    return null
  }

  switch (linkType) {
    case 'tel':
      return `tel:${value}`
    case 'mailto':
      return `mailto:${value}`
    case 'url':
      return value.startsWith('https') ? value : null
  }
}

export function LinkedValue({
  value,
  linkType,
}: {
  value: string
  linkType: LinkType
}) {
  const href = getLinkHref(value, linkType)
  if (!href) return <>{value}</>

  return (
    <a
      href={href}
      className="text-primary hover:underline"
      {...(linkType === 'url' ? { target: '_blank', rel: 'noreferrer' } : {})}
    >
      {value}
    </a>
  )
}
