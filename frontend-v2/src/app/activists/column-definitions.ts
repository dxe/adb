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
  | 'Advanced'

export interface ColumnDefinition {
  name: ActivistColumnName
  label: string
  category: ColumnCategory
  isDate?: boolean // If true, format string values as dates
  hidden?: boolean // If true, hide from user-facing column selectors
}

export const DEFAULT_COLUMNS: ActivistColumnName[] = [
  'name',
  'email',
  'phone',
  'activist_level',
]

export const COLUMN_DEFINITIONS: ColumnDefinition[] = [
  // Chapter (conditionally shown based on filters)
  { name: 'chapter_name', label: 'Chapter', category: 'Basic Info' },

  // Basic Info
  { name: 'name', label: 'Name', category: 'Basic Info' },
  { name: 'preferred_name', label: 'Preferred Name', category: 'Basic Info' },
  { name: 'pronouns', label: 'Pronouns', category: 'Basic Info' },
  { name: 'email', label: 'Email', category: 'Basic Info' },
  { name: 'phone', label: 'Phone', category: 'Basic Info' },
  { name: 'facebook', label: 'Facebook', category: 'Basic Info' },
  { name: 'activist_level', label: 'Level', category: 'Basic Info' },
  { name: 'dob', label: 'Birthday', category: 'Basic Info', isDate: true },
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
  {
    name: 'first_event',
    label: 'First Event',
    category: 'Event Attendance',
    isDate: true,
  },
  {
    name: 'first_event_name',
    label: 'First Event Name',
    category: 'Event Attendance',
  },
  {
    name: 'last_event',
    label: 'Last Event',
    category: 'Event Attendance',
    isDate: true,
  },
  {
    name: 'last_event_name',
    label: 'Last Event Name',
    category: 'Event Attendance',
  },
  {
    name: 'last_action',
    label: 'Last Action',
    category: 'Event Attendance',
    isDate: true,
  },
  {
    name: 'months_since_last_action',
    label: 'Mo. Since Last Action',
    category: 'Event Attendance',
  },
  {
    name: 'total_points',
    label: 'Points',
    category: 'Event Attendance',
  },
  { name: 'active', label: 'Active', category: 'Event Attendance' },
  { name: 'status', label: 'Status', category: 'Event Attendance' },
  {
    name: 'mpp_requirements',
    label: 'DA&C Current Month',
    category: 'Event Attendance',
  },
  { name: 'mpi', label: 'MPI', category: 'Event Attendance' },

  // Trainings
  { name: 'training0', label: 'Workshop', category: 'Trainings' },
  { name: 'training1', label: 'Consent & Oppression', category: 'Trainings' },
  {
    name: 'training4',
    label: 'Building Purposeful Communities',
    category: 'Trainings',
  },
  {
    name: 'training5',
    label: 'Leadership & Management',
    category: 'Trainings',
  },
  { name: 'training6', label: 'Vision and Strategy', category: 'Trainings' },
  { name: 'training_protest', label: 'Tier 2 Protest', category: 'Trainings' },
  {
    name: 'consent_quiz',
    label: 'Consent Refresher Quiz',
    category: 'Trainings',
  },

  // Application Info
  {
    name: 'dev_application_date',
    label: 'Applied',
    category: 'Application Info',
    isDate: true,
  },
  {
    name: 'dev_application_type',
    label: 'Application Type',
    category: 'Application Info',
  },
  {
    name: 'prospect_chapter_member',
    label: 'Prospective Chapter Member',
    category: 'Application Info',
  },
  {
    name: 'prospect_organizer',
    label: 'Prospective Organizer',
    category: 'Application Info',
  },

  // Circle Info
  { name: 'geo_circles', label: 'Geo-Circle', category: 'Circle Info' },

  // Prospect Info
  {
    name: 'assigned_to',
    label: 'Assigned To ID',
    category: 'Prospect Info',
    hidden: true,
  },
  {
    name: 'assigned_to_name',
    label: 'Assigned To',
    category: 'Prospect Info',
  },
  {
    name: 'followup_date',
    label: 'Follow-up Date',
    category: 'Prospect Info',
    isDate: true,
  },
  {
    name: 'total_interactions',
    label: 'Interactions',
    category: 'Prospect Info',
  },
  {
    name: 'last_interaction_date',
    label: 'Last Interaction',
    category: 'Prospect Info',
    isDate: true,
  },

  // Referral Info
  { name: 'source', label: 'Source', category: 'Referral Info' },
  {
    name: 'interest_date',
    label: 'Interest Date',
    category: 'Referral Info',
    isDate: true,
  },
  { name: 'referral_friends', label: 'Close Ties', category: 'Referral Info' },
  { name: 'referral_apply', label: 'Referral', category: 'Referral Info' },
  {
    name: 'referral_outlet',
    label: 'Referral Outlet',
    category: 'Referral Info',
  },

  // Development
  { name: 'dev_quiz', label: 'Quiz', category: 'Development' },
  { name: 'dev_interest', label: 'Interests', category: 'Development' },
  { name: 'connector', label: 'Coach', category: 'Development' },
  {
    name: 'last_connection',
    label: 'Last Coaching',
    category: 'Development',
    isDate: true,
  },

  // Chapter Membership
  {
    name: 'cm_first_email',
    label: 'First Text',
    category: 'Chapter Membership',
    isDate: true,
  },
  {
    name: 'cm_approval_email',
    label: 'Approval Email',
    category: 'Chapter Membership',
    isDate: true,
  },
  { name: 'vision_wall', label: 'Vision Wall', category: 'Chapter Membership' },
  {
    name: 'voting_agreement',
    label: 'Voting Agreement',
    category: 'Chapter Membership',
  },

  // Other
  { name: 'notes', label: 'Notes', category: 'Other' },
  { name: 'hiatus', label: 'Hiatus', category: 'Other' },

  // Developer
  { name: 'id', label: 'ID', category: 'Advanced' },
]

export const groupColumnsByCategory = () => {
  const grouped = new Map<ColumnCategory, ColumnDefinition[]>()

  COLUMN_DEFINITIONS.forEach((col) => {
    const existing = grouped.get(col.category) || []
    grouped.set(col.category, [...existing, col])
  })

  return grouped
}

// Normalizes column selection. Ensures columns:
//  * include required columns
//  * are not duplicated
//  * are sorted according to their order in COLUMN_DEFINITIONS
export const normalizeColumns = (
  columns: ActivistColumnName[],
): ActivistColumnName[] => {
  // Deduplicate columns
  const uniqueColumns = Array.from(new Set(columns))

  // Add required columns
  if (!uniqueColumns.includes('name')) {
    uniqueColumns.push('name')
  }

  // Sort according to COLUMN_DEFINITIONS order
  const orderMap = new Map<ActivistColumnName, number>()
  COLUMN_DEFINITIONS.forEach((col, index) => {
    orderMap.set(col.name, index)
  })

  return uniqueColumns.sort((a, b) => {
    const orderA = orderMap.get(a) ?? Number.MAX_SAFE_INTEGER
    const orderB = orderMap.get(b) ?? Number.MAX_SAFE_INTEGER
    return orderA - orderB
  })
}

// Normalizes columns and ensures chapter_name is present if and only if
// searching across chapters. This centralizes the filter-dependent column logic.
export const normalizeColumnsForFilters = (
  columns: ActivistColumnName[],
  searchAcrossChapters: boolean,
): ActivistColumnName[] => {
  let adjustedColumns = [...columns]

  // Ensure chapter_name is visible if and only if searching across chapters
  if (searchAcrossChapters) {
    if (!adjustedColumns.includes('chapter_name')) {
      adjustedColumns.unshift('chapter_name')
    }
  } else {
    if (adjustedColumns.includes('chapter_name')) {
      adjustedColumns = adjustedColumns.filter((col) => col !== 'chapter_name')
    }
  }

  return normalizeColumns(adjustedColumns)
}
