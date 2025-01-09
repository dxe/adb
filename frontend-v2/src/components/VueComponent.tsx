import Script from "next/script";

export const VueComponent = ({
  name,
  data,
}: {
  name: string;
  data?: { [k: string]: string };
}) => {
  return (
    <>
      <div
        id="app"
        dangerouslySetInnerHTML={{
          __html: `
            <${name} ${Object.entries(data ?? {})
            .map(([k, v]) => `${k}="${v}"`)
            .join(" ")}></${name}>
            `,
        }}
      />
      {/* TODO(jh): this needs static resource hash */}
      <Script src="/dist/adb.js" strategy="afterInteractive" />
    </>
  );
};
