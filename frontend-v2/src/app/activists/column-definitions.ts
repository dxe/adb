import { ActivistColumnName } from '@/lib/api'

export type ColumnCategory =
  | 'Basic Info'
  | 'Location'
  | 'Event Attendance'
  | 'Trainings'
  | 'Application Info'
  | 'Circle Info'
  | 'Prospect Info'
  | 'Referral Info'
  | 'Development'
  | 'Chapter Membership'
  | 'Other'

export interface ColumnDefinition {
  name: ActivistColumnName
  label: string
  category: ColumnCategory
}

export const COLUMN_DEFINITIONS: ColumnDefinition[] = [
  // Basic Info
  { name: 'name', label: 'Name', category: 'Basic Info' },
  { name: 'preferred_name', label: 'Preferred Name', category: 'Basic Info' },
  { name: 'pronouns', label: 'Pronouns', category: 'Basic Info' },
  { name: 'email', label: 'Email', category: 'Basic Info' },
  { name: 'phone', label: 'Phone', category: 'Basic Info' },
  { name: 'facebook', label: 'Facebook', category: 'Basic Info' },
  { name: 'activist_level', label: 'Level', category: 'Basic Info' },
  { name: 'dob', label: 'Birthday', category: 'Basic Info' },
  { name: 'accessibility', label: 'Accessibility', category: 'Basic Info' },
  { name: 'language', label: 'Language', category: 'Basic Info' },

  // Location
  { name: 'location', label: 'Location', category: 'Location' },
  { name: 'street_address', label: 'Street Address', category: 'Location' },
  { name: 'city', label: 'City', category: 'Location' },
  { name: 'state', label: 'State', category: 'Location' },
  { name: 'lat', label: 'Latitude', category: 'Location' },
  { name: 'lng', label: 'Longitude', category: 'Location' },

  // Event Attendance
  { name: 'total_events', label: 'Total Events', category: 'Event Attendance' },
  { name: 'first_event', label: 'First Event', category: 'Event Attendance' },
  { name: 'first_event_name', label: 'First Event Name', category: 'Event Attendance' },
  { name: 'last_event', label: 'Last Event', category: 'Event Attendance' },
  { name: 'last_event_name', label: 'Last Event Name', category: 'Event Attendance' },
  { name: 'mpi', label: 'MPI', category: 'Event Attendance' },

  // Trainings
  { name: 'training0', label: 'Workshop', category: 'Trainings' },
  { name: 'training1', label: 'Consent & Oppression', category: 'Trainings' },
  { name: 'training4', label: 'Building Purposeful Communities', category: 'Trainings' },
  { name: 'training5', label: 'Leadership & Management', category: 'Trainings' },
  { name: 'training6', label: 'Vision and Strategy', category: 'Trainings' },
  { name: 'training_protest', label: 'Tier 2 Protest', category: 'Trainings' },
  { name: 'consent_quiz', label: 'Consent Refresher Quiz', category: 'Trainings' },

  // Application Info
  { name: 'dev_application_date', label: 'Applied', category: 'Application Info' },
  { name: 'dev_application_type', label: 'Application Type', category: 'Application Info' },
  { name: 'prospect_chapter_member', label: 'Prospective Chapter Member', category: 'Application Info' },
  { name: 'prospect_organizer', label: 'Prospective Organizer', category: 'Application Info' },

  // Circle Info
  // Note: geo_circles not yet implemented in backend

  // Prospect Info
  { name: 'assigned_to', label: 'Assigned To', category: 'Prospect Info' },
  { name: 'followup_date', label: 'Follow-up Date', category: 'Prospect Info' },
  // Note: total_interactions, last_interaction_date not yet implemented in backend

  // Referral Info
  { name: 'source', label: 'Source', category: 'Referral Info' },
  { name: 'interest_date', label: 'Interest Date', category: 'Referral Info' },
  { name: 'referral_friends', label: 'Close Ties', category: 'Referral Info' },
  { name: 'referral_apply', label: 'Referral', category: 'Referral Info' },
  { name: 'referral_outlet', label: 'Referral Outlet', category: 'Referral Info' },

  // Development
  { name: 'dev_quiz', label: 'Quiz', category: 'Development' },
  { name: 'dev_interest', label: 'Interests', category: 'Development' },
  { name: 'connector', label: 'Coach', category: 'Development' },
  // Note: last_connection not yet implemented in backend

  // Chapter Membership
  { name: 'cm_first_email', label: 'First Text', category: 'Chapter Membership' },
  { name: 'cm_approval_email', label: 'Approval Email', category: 'Chapter Membership' },
  { name: 'vision_wall', label: 'Vision Wall', category: 'Chapter Membership' },
  { name: 'voting_agreement', label: 'Voting Agreement', category: 'Chapter Membership' },
  { name: 'mpp_requirements', label: 'MPP Requirements', category: 'Chapter Membership' },

  // Other
  { name: 'notes', label: 'Notes', category: 'Other' },
  { name: 'hiatus', label: 'Hiatus', category: 'Other' },

  // Note: chapter_id and chapter_name are handled separately based on filter state
]

// Group columns by category
export const groupColumnsByCategory = () => {
  const grouped = new Map<ColumnCategory, ColumnDefinition[]>()

  COLUMN_DEFINITIONS.forEach((col) => {
    const existing = grouped.get(col.category) || []
    grouped.set(col.category, [...existing, col])
  })

  return grouped
}

// Get default columns based on whether showing all chapters
export const getDefaultColumns = (showAllChapters: boolean): ActivistColumnName[] => {
  const baseColumns: ActivistColumnName[] = ['name', 'email', 'phone', 'activist_level']

  if (showAllChapters) {
    return ['chapter_name', ...baseColumns]
  }

  return baseColumns
}

// Sort columns according to their order in COLUMN_DEFINITIONS
// Also ensures uniqueness and that 'name' is always included
export const sortColumnsByDefinitionOrder = (
  columns: ActivistColumnName[],
): ActivistColumnName[] => {
  // Deduplicate columns
  const uniqueColumns = Array.from(new Set(columns))

  // Ensure 'name' is always included
  if (!uniqueColumns.includes('name')) {
    uniqueColumns.push('name')
  }

  // Create a map of column name to its index in COLUMN_DEFINITIONS
  const orderMap = new Map<ActivistColumnName, number>()
  COLUMN_DEFINITIONS.forEach((col, index) => {
    orderMap.set(col.name, index)
  })

  // Special handling for chapter_name (should always be first if present)
  const hasChapterName = uniqueColumns.includes('chapter_name')
  const columnsWithoutChapterName = uniqueColumns.filter((col) => col !== 'chapter_name')

  // Sort columns based on their position in COLUMN_DEFINITIONS
  const sorted = columnsWithoutChapterName.sort((a, b) => {
    const orderA = orderMap.get(a) ?? Number.MAX_SAFE_INTEGER
    const orderB = orderMap.get(b) ?? Number.MAX_SAFE_INTEGER
    return orderA - orderB
  })

  // Add chapter_name back at the beginning if it was present
  if (hasChapterName) {
    return ['chapter_name', ...sorted]
  }

  return sorted
}
