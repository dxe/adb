import { parseAsStringLiteral } from 'nuqs/server'

// This page views/edits both "Interest Circles" and "Geo-Circles".
export const CIRCLE_MODES = ['interest', 'geo'] as const
export type CircleMode = (typeof CIRCLE_MODES)[number]

export const circleSearchParamParsers = {
  type: parseAsStringLiteral(CIRCLE_MODES).withDefault('interest'),
}
