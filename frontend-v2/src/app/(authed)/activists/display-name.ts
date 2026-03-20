import type { ActivistJSON } from '@/lib/api'

export function getActivistDisplayName(activist: ActivistJSON) {
  const hasName = activist.name && activist.name.trim() !== ''
  return {
    text: hasName ? activist.name : `<Activist ID: ${activist.id}>`,
    isPlaceholder: !hasName,
  }
}
