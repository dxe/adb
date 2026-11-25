# navigation

The menu is defined in `nav-data.ts`: dropdown groups, each with link items.

## adding an item

- `href` includes the `/v2` prefix, unlike `<Link href>` in the rest of the app. Hrefs that start with `/v2` render as `<Link>` with the prefix removed. Any other href renders as a plain `<a>` and does a full page load.
- For saved activist filters, build the href with `activistsHref(...)` and set `page: 'ActivistListV2'`. That value makes the item highlight only when the URL's query matches the filter exactly, and makes a click on `/activists` update the filters in place instead of navigating. Nav code doesn't read any other `page` value.
- If you leave out `visibleForNonSFBay`, the item uses its group's value. Leave out `roleRequired` unless the item needs a stricter role than its group. See the rules below.
- Run `pnpm exec vitest run src/components/nav` after editing `nav-data.ts`.

## who sees an item

`evaluateNavAccess` in `nav-access.ts` checks each group, then each item in a group the user can see. An item shows only if it passes both checks below.

### region: `visibleForNonSFBay`

- If an item doesn't set it, it uses its group's value. If neither sets it, the default is `false`, so the item shows only in SF Bay.
- An item can't show if its group is hidden. Setting `true` on an item in an SF Bay-only group does nothing.
- The check uses the active chapter (`user.ChapterID`), so admins don't skip it.

### roles: `roleRequired`

- The user needs any one of the listed roles. Leaving it out means anyone logged in passes.
- Items don't inherit it from their group. The group's roles still apply in practice, because a hidden group hides all its items.
- `admin`, `organizer` and `attendance` are tiers, where each one includes the ones below it. `admin` passes all three, `organizer` passes `organizer` and `attendance`.
- Other roles, such as `intl_coordinator`, aren't tiered: only that exact role passes, and admins don't. To let admins in too, list both, e.g. `['admin', 'intl_coordinator']`.

### this is not authorization

Hiding a nav item doesn't stop anyone opening its URL. The page and the Go API must check access themselves, so make sure they match the item's `roleRequired`.
