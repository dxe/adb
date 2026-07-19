import { parseAsStringLiteral } from 'nuqs/server'

// This single page serves both the "Interest Circles" and "Geo-Circles" nav
// entries (mirroring the legacy Vue `CirclesList.vue` component, which is
// mounted once per Go template with a different `title` prop). The mode is
// selected via the `type` query param instead.
export const CIRCLE_MODES = ['interest', 'geo'] as const
export type CircleMode = (typeof CIRCLE_MODES)[number]

export const circleSearchParamParsers = {
  type: parseAsStringLiteral(CIRCLE_MODES).withDefault('interest'),
}
