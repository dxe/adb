import Script from 'next/script'
import { fetchStaticResourceHash } from 'app/static-resource-hash'
import { fetchSession } from 'app/session'
import { getAuthCookies } from 'lib/auth'

// Allows the Vue AdbNav component to be used within the React app
// so that the UI is more consistent. Once most pages are rebuilt in
// React, we should migrate the Navbar to React as well.
// The biggest downside currently is that the entire app has
// to reload whenever a link is clicked, due to not using the <Link>
// component to navigate between pages.
//
// Works best with 'has-navbar-fixed-top' class on the body tag to prevent
// layout shift, since the Vue script is loaded asynchronously. Without this,
// the navbar pushes remaining page content down once it finally loads.
export const VueNavbar = async (props: {
  /** The name of the active page, corresponding to the name in Vue. */
  pageName: string
}) => {
  const session = await fetchSession(await getAuthCookies())
  const staticResourceHash = await fetchStaticResourceHash()

  return (
    <>
      <link
        rel="stylesheet"
        type="text/css"
        href="/static/external/buefy.min.css"
      />
      <div
        id="app"
        className="shadow-none"
        dangerouslySetInnerHTML={{
          __html: `
          <adb-nav
            page="${props.pageName}"
            user="${session?.user?.Name ?? ''}"
            role="${session?.user?.role ?? ''}"
            chapter="${session?.user?.ChapterName ?? ''}">
          </adb-nav>
        `,
        }}
        suppressHydrationWarning
      />
      <Script src={`/dist/adb.js?hash=${staticResourceHash}`} />
    </>
  )
}
