import Script from 'next/script'
import { fetchStaticResourceHash } from 'app/static-resource-hash'
import { fetchSession } from 'app/session'
import { getCookies } from 'lib/auth'

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
  const [session, staticResourceHash] = await Promise.all([
    fetchSession(await getCookies()),
    fetchStaticResourceHash(),
  ])

  return (
    <>
      <link
        rel="stylesheet"
        type="text/css"
        href="/static/external/buefy.min.css"
      />
      {/* Show a white background where the navbar will appear to reduce the
      effect of the navbar flashing in and out as the user navigates between
      pages. */}
      <div
        style={{
          position: 'fixed' /* Keep at the top on scroll */,
          top: 0 /* Aligns to the top of the screen */,
          left: 0,
          width: '100%',
          height: '3.25rem' /* Same height as Buefy navbar */,
          backgroundColor: '#fff',
          zIndex: 15 /* Place behind actual navbar which is z-index 30 */,
        }}
      ></div>
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
