import { useSession } from "@/hooks/use-session";
import Script from "next/script";
import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import ky from "ky";
import { z } from "zod";

const StaticResourcesHashResp = z.object({
  hash: z.string(),
});

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

  // TODO(jh): prob should do this elsewhere b/c a logged-out navbar is fine & should show a 'login' button.
  useEffect(() => {
    if (session.isLoading) {
      return;
    }
    if (!session.user) {
      window.location.pathname = "/login";
    }
  }, [session.isLoading, session.user]);

  const { data: staticResourceHash } = useQuery({
    queryKey: ["static_resources_hash"],
    queryFn: async () => {
      const resp = await ky.get("/static_resources_hash").json();
      return StaticResourcesHashResp.parse(resp);
    },
  });

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
        src={`/dist/adb.js?hash=${staticResourceHash.hash}`}
        strategy="afterInteractive"
      />
    </>
  );
};
