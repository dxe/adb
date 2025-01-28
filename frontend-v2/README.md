This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://nextjs.org/docs/pages/api-reference/create-next-app).

## Getting Started

See the main project README for initial, one-time setup as well as instructions
to build the project and run the development servers.

## Frontend migration

We have two frontend applications that are built independently except that the
React app imports a single component from the Vue app--the navigation bar.
This allows us to define in one place which pages belong to which application.

Each application is hosted at the same origin, so cookies are shared.

All React pages are served under the path `v2/*`  (via a proxy in local dev so
the next.js dev server can properly run, and in prod so that Next.js can do SSR
for smoother page loads). The navigation bar links to a mix of `v2/` and non-v2
destinations, which determines which app will load as the user navigates.

To upgrade a page, one implements it in the React app, changes the link in the
navigation to go to the `v2/` page, then deletes the Vue page.

Authentication is handled in vanilla js, not within Vue/React.

## Learn about Next.js

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn-pages-router) - an interactive Next.js tutorial.

