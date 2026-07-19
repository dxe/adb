import { parseAsStringLiteral } from 'nuqs/server'

// One page serves both the "Interest Circles" and "Geo-Circles" nav entries;
// `type` replaces the `title` prop the legacy Vue `CirclesList.vue` was mounted with.
export const CIRCLE_MODES = ['interest', 'geo'] as const
export type CircleMode = (typeof CIRCLE_MODES)[number]

export const circleSearchParamParsers = {
  type: parseAsStringLiteral(CIRCLE_MODES).withDefault('interest'),
}
