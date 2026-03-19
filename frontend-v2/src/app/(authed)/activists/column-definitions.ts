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
  description?: string // Optional help text shown in column selector tooltip
  category: ColumnCategory
  isDate?: boolean // If true, format string values as dates
  hidden?: boolean // If true, hide from user-facing column selectors
  defaultWidth?: number // Default column width in pixels (default: 150)
  minWidth?: number // Minimum column width in pixels (default: 60)
  hideOnDetailPage?: boolean // If true, omit from the individual activist detail view
  linkType?: 'tel' | 'url' | 'mailto' // If set, render as a link on the detail page
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
  {
    name: 'preferred_name',
    label: 'Preferred Name',
    category: 'Basic Info',
    description: 'First name or nickname, may be used for SMS',
    defaultWidth: 90,
  },
  {
    name: 'pronouns',
    label: 'Pronouns',
    category: 'Basic Info',
    defaultWidth: 50,
  },
  { name: 'email', label: 'Email', category: 'Basic Info', linkType: 'mailto' },
  {
    name: 'phone',
    label: 'Phone',
    category: 'Basic Info',
    defaultWidth: 100,
    linkType: 'tel',
  },
  {
    name: 'facebook',
    label: 'Facebook',
    category: 'Basic Info',
    linkType: 'url',
  },
  {
    name: 'activist_level',
    label: 'Level',
    category: 'Basic Info',
    defaultWidth: 140,
  },
  {
    name: 'dob',
    label: 'Birthday',
    category: 'Basic Info',
    isDate: true,
    defaultWidth: 100,
  },
  {
    name: 'accessibility',
    label: 'Accessibility',
    category: 'Basic Info',
    defaultWidth: 50,
  },
  {
    name: 'language',
    label: 'Language',
    category: 'Basic Info',
    defaultWidth: 50,
  },

  // Location
  { name: 'location', label: 'Location', category: 'Location' },
  {
    name: 'street_address',
    label: 'Street Address',
    category: 'Location',
    defaultWidth: 100,
  },
  {
    name: 'city',
    label: 'City',
    category: 'Location',
    defaultWidth: 100,
  },
  {
    name: 'state',
    label: 'State',
    category: 'Location',
    defaultWidth: 100,
  },
  {
    name: 'lat',
    label: 'Latitude',
    category: 'Location',
    hideOnDetailPage: true,
  },
  {
    name: 'lng',
    label: 'Longitude',
    category: 'Location',
    hideOnDetailPage: true,
  },

  // Event Attendance
  {
    name: 'total_events',
    label: 'Total Events',
    category: 'Event Attendance',
    defaultWidth: 90,
  },
  {
    name: 'first_event',
    label: 'First Event Date',
    category: 'Event Attendance',
    isDate: true,
    defaultWidth: 200,
  },
  {
    name: 'first_event_name',
    label: 'First Event Name',
    category: 'Event Attendance',
    defaultWidth: 200,
  },
  {
    name: 'last_event',
    label: 'Last Event Date',
    category: 'Event Attendance',
    isDate: true,
    defaultWidth: 200,
  },
  {
    name: 'last_event_name',
    label: 'Last Event Name',
    category: 'Event Attendance',
    defaultWidth: 200,
  },
  {
    name: 'last_action',
    label: 'Last Action Date',
    category: 'Event Attendance',
    isDate: true,
  },
  {
    name: 'months_since_last_action',
    label: 'Months Since Last Action',
    category: 'Event Attendance',
    defaultWidth: 50,
  },
  {
    name: 'total_points',
    label: 'Leadership points',
    category: 'Event Attendance',
    defaultWidth: 50,
  },
  { name: 'active', label: 'Active', category: 'Event Attendance' },
  { name: 'status', label: 'Status', category: 'Event Attendance' },
  {
    name: 'mpp_requirements',
    label: 'DA&C Current Month',
    category: 'Event Attendance',
    defaultWidth: 80,
    description:
      'Attended both a Direct Action & a Community event in current month (historical requirement for MPP)',
  },
  {
    name: 'mpi',
    label: 'MPI',
    category: 'Event Attendance',
    defaultWidth: 30,
    description:
      'Whether the activist satisfied the MPP in either the current month or the previous month (see dxe.io/mpp)',
  },

  // Trainings
  {
    name: 'training0',
    label: 'Workshop',
    category: 'Trainings',
    defaultWidth: 100,
  },
  {
    name: 'training1',
    label: 'Consent & Oppression',
    category: 'Trainings',
    defaultWidth: 100,
  },
  {
    name: 'training4',
    label: 'Building Purposeful Communities',
    category: 'Trainings',
    defaultWidth: 100,
  },
  {
    name: 'training5',
    label: 'Leadership & Management',
    category: 'Trainings',
    defaultWidth: 100,
  },
  {
    name: 'training6',
    label: 'Vision and Strategy',
    category: 'Trainings',
    defaultWidth: 100,
  },
  {
    name: 'training_protest',
    label: 'Tier 2 Protest',
    category: 'Trainings',
    defaultWidth: 100,
  },
  {
    name: 'consent_quiz',
    label: 'Consent Refresher Quiz',
    category: 'Trainings',
    defaultWidth: 100,
  },

  // Application Info
  {
    name: 'dev_application_date',
    label: 'Applied',
    category: 'Application Info',
    isDate: true,
    defaultWidth: 100,
  },
  {
    name: 'dev_application_type',
    label: 'Application Type',
    category: 'Application Info',
    defaultWidth: 80,
  },
  {
    name: 'prospect_chapter_member',
    label: 'Prospective Chapter Member',
    category: 'Application Info',
    defaultWidth: 60,
  },
  {
    name: 'prospect_organizer',
    label: 'Prospective Organizer',
    category: 'Application Info',
    defaultWidth: 100,
  },

  // Circle Info
  {
    name: 'geo_circles',
    label: 'Geo-Circle',
    category: 'Circle Info',
    defaultWidth: 100,
  },

  // Prospect Info
  {
    name: 'assigned_to',
    label: 'Assigned To ID',
    category: 'Prospect Info',
    hidden: true,
    hideOnDetailPage: true,
  },
  {
    name: 'assigned_to_name',
    label: 'Assigned To',
    category: 'Prospect Info',
    defaultWidth: 100,
  },
  {
    name: 'followup_date',
    label: 'Follow-up Date',
    category: 'Prospect Info',
    isDate: true,
    defaultWidth: 110,
    description: 'Date to follow up',
  },
  {
    name: 'total_interactions',
    label: 'Interactions',
    category: 'Prospect Info',
    defaultWidth: 80,
  },
  {
    name: 'last_interaction_date',
    label: 'Last Interaction',
    category: 'Prospect Info',
    isDate: true,
    defaultWidth: 100,
  },

  // Referral Info
  {
    name: 'source',
    label: 'Source',
    category: 'Referral Info',
    defaultWidth: 100,
  },
  {
    name: 'interest_date',
    label: 'Interest Date',
    category: 'Referral Info',
    isDate: true,
    defaultWidth: 100,
    description: 'Date activist submitted the interest form',
  },
  {
    name: 'referral_friends',
    label: 'Close Ties',
    category: 'Referral Info',
    defaultWidth: 100,
  },
  {
    name: 'referral_apply',
    label: 'Referral',
    category: 'Referral Info',
    defaultWidth: 100,
    description: '"Who encouraged you to sign up?"',
  },
  {
    name: 'referral_outlet',
    label: 'Referral Outlet',
    category: 'Referral Info',
    defaultWidth: 100,
    description: '"How did you hear about DxE?"',
  },

  // Development
  {
    name: 'dev_quiz',
    label: 'Quiz',
    category: 'Development',
    defaultWidth: 100,
  },
  {
    name: 'dev_interest',
    label: 'Interests',
    category: 'Development',
    defaultWidth: 100,
  },
  {
    name: 'connector',
    label: 'Coach',
    category: 'Development',
    defaultWidth: 125,
  },
  {
    name: 'last_connection',
    label: 'Last Coaching',
    category: 'Development',
    isDate: true,
    defaultWidth: 100,
  },

  // Chapter Membership
  {
    name: 'cm_first_email',
    label: 'First Text',
    category: 'Chapter Membership',
    isDate: true,
    defaultWidth: 100,
    description: 'Date of first SMS message sent to activist',
  },
  {
    name: 'cm_approval_email',
    label: 'Approval Email',
    category: 'Chapter Membership',
    isDate: true,
    defaultWidth: 100,
  },
  {
    name: 'vision_wall',
    label: 'Vision Wall',
    category: 'Chapter Membership',
    defaultWidth: 80,
  },
  {
    name: 'voting_agreement',
    label: 'Voting Agreement',
    category: 'Chapter Membership',
    defaultWidth: 50,
  },

  // Other
  { name: 'notes', label: 'Notes', category: 'Other', defaultWidth: 100 },
  { name: 'hiatus', label: 'Hiatus', category: 'Other', defaultWidth: 50 },

  // Developer
  {
    name: 'id',
    label: 'ID',
    category: 'Advanced',
    description: 'Database ID of the activist',
    defaultWidth: 50,
    hideOnDetailPage: true,
  },
]

export const COLUMN_DEFINITION_BY_NAME = Object.fromEntries(
  COLUMN_DEFINITIONS.map((definition) => [definition.name, definition]),
) as Record<ActivistColumnName, ColumnDefinition>

export const COLUMN_ORDER_BY_NAME = new Map<ActivistColumnName, number>(
  COLUMN_DEFINITIONS.map((column, index) => [column.name, index]),
)

export const GROUPED_COLUMNS_BY_CATEGORY = (() => {
  const grouped = new Map<ColumnCategory, ColumnDefinition[]>()

  COLUMN_DEFINITIONS.forEach((column) => {
    const existing = grouped.get(column.category) || []
    grouped.set(column.category, [...existing, column])
  })

  return grouped
})()

export const groupColumnsByCategory = () => GROUPED_COLUMNS_BY_CATEGORY

export function isActivistColumnName(
  value: string,
): value is ActivistColumnName {
  return Object.hasOwn(COLUMN_DEFINITION_BY_NAME, value)
}
