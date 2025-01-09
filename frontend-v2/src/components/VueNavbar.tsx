import { useSession } from "@/hooks/use-session";
import Script from "next/script";
import { useStaticResourceHash } from "@/hooks/use-static-resource-hash";

// Allows the Vue AdbNav component to be used within the React app
// for more seamless UX. Once most pages are rebuilt in
// React, we should migrate the Navbar to React as well.
// The biggest downside currently is that the entire app has
// to reload whenever a link is clicked, due to not using the <Link>
// component to navigate between pages.
export const VueNavbar = (props: {
  /** The name of the active page, corresponding to the name in Vue. */
  pageName: string;
}) => {
  const session = useSession();
  const staticResourceHash = useStaticResourceHash();

  return !staticResourceHash || session.isLoading ? null : (
    <>
      {/* eslint-disable-next-line @next/next/no-css-tags */}
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
                user="${session.user?.Name}"
                role="${session?.user.role}"
                chapter="${session.user.ChapterName}">
              </adb-nav>
            `,
        }}
      />
      <Script
        src={`/dist/adb.js?hash=${staticResourceHash}`}
        strategy="afterInteractive"
      />
    </>
  );
};
