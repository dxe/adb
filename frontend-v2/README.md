This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://nextjs.org/docs/pages/api-reference/create-next-app).

## Getting Started

See the main project README for initial, one-time setup as well as instructions
to build the project and run the development servers.

## Frontend migration

We have two frontend applications that are built independently except that the
React app imports a single component from the Vue app--the navigation bar.
This allows us to define in one place which pages belong to which application.

Each application is hosted at the same origin, so cookies are shared.

All React pages are served under the path `v2/*` (via a proxy in local dev so
the next.js dev server can properly run, and in prod so that Next.js can do SSR
for smoother page loads). The navigation bar links to a mix of `v2/` and non-v2
destinations, which determines which app will load as the user navigates.

To upgrade a page, one implements it in the React app, changes the link in the
navigation to go to the `v2/` page, then deletes the Vue page.

Authentication is handled in vanilla js, not within Vue/React.

## Learn about Next.js

See [Next.js Documentation][next] to learn more about Next.js.

[next]: https://nextjs.org/docs

### App Router

This project uses Next.js's new App Router rather than the Pages Router. Many
topics in Next.js documentation are written about assuming use of a specific
router and have equivalent pages for the other router, so be sure to make sure
you are reading the right version of the docs. As of Feb 2025, there is a
drop-down in the upper left of Next.js documentation allowing you to toggle
between the choice of routers.

## Server Components

Next.js App Router uses [React Server Components][rsc] for server rendering that works
not only when loading the initial page but also while navigating within the app.

[rsc]: https://react.dev/reference/rsc/server-components

## TanStack React Query

This project uses TanStack React Query to make data fetching and mutation
easier. Our usage of React Query is specific to Next.js App Router and
React Server Components. See React Query's documentation for this in their
[Advanced SSR guide][rq-ssr]

[rq-ssr]: https://tanstack.com/query/latest/docs/framework/react/guides/advanced-ssr

# Build

## Testing in production

If you have AWS access, you can deploy changes to the v2 frontend in production
before merging changes to the main branch. As actual traffic to v2 pages
increases, this will become a riskier testing method.

First, build the images, assign the right tags for use in the AWS ECR repo,
log into the AWS CLI and then push to the ECR repo. Watchtower running in EC2
will automatically pick up the changes and make them live.

```bash
make prod_build
docker tag dxe/adb-next 521324062467.dkr.ecr.us-west-2.amazonaws.com/dxe/adb-next
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 521324062467.dkr.ecr.us-west-2.amazonaws.com
docker push 521324062467.dkr.ecr.us-west-2.amazonaws.com/dxe/adb-next:latest
```
