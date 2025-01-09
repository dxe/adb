import { Html, Head, Main, NextScript } from "next/document";
import Script from "next/script";

export default function Document() {
  return (
    <Html lang="en">
      <Head>
        <meta charSet="utf-8" />
        <meta http-equiv="X-UA-Compatible" content="IE=edge" />
        <meta
          name="viewport"
          content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no"
        />
        {/* TODO(jh): need to provide csrf token here */}
        <meta name="csrf-token" content="TODO" />
        {/* eslint-disable-next-line @next/next/no-title-in-document-head */}
        <title>Activist Database</title>
        <link rel="icon" type="image/png" href="/static/img/favicon.png" />
        {/* TODO(jh): Remove this once we no longer use Vue. */}
        <link
          rel="stylesheet"
          href="https://cdn.jsdelivr.net/npm/@mdi/font@5.8.55/css/materialdesignicons.min.css"
        />
        {/* TODO(jh): Remove this once we no longer use Vue. */}
        {/* eslint-disable-next-line @next/next/no-css-tags */}
        <link
          rel="stylesheet"
          type="text/css"
          // TODO(jh): needs static resource hash
          href="/static/css/style.css"
        />
        {/* TODO(jh): Remove this once we no longer use Vue. */}
        {/* eslint-disable-next-line @next/next/no-css-tags */}
        <link
          rel="stylesheet"
          type="text/css"
          href="/static/external/buefy.min.css"
        />
        <Script src="/static/external/jquery-3.2.1.js" />
      </Head>
      <body className="antialiased">
        <Main />
        <NextScript />
      </body>
    </Html>
  );
}
