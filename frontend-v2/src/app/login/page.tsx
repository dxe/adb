import {
  dehydrate,
  HydrationBoundary,
  QueryClient,
} from '@tanstack/react-query'
import { ContentWrapper } from '@/app/content-wrapper'
import { getCookies } from '@/lib/auth'
import { fetchSession } from '@/app/session'
import { JSX } from 'react'
import Script from 'next/script'

export default async function LoginPage() {
  const queryClient = new QueryClient()
  const session = await fetchSession(await getCookies())

  let userJsx: JSX.Element | null = null
  if (session.user) {
    userJsx = <p>Logged in as {session.user.Email}</p>
  } else {
    userJsx = <p>Not logged in.</p>
  }

  // This HTML is generated from:
  // https://developers.google.com/identity/gsi/web/tools/configurator
  //
  // This invokes the Sign-in-with-Google HTML API. Note that Google also provides a JavaScript API for more complex
  // use cases.
  let signInwithGoogleHtmlInvocation = (
    <>
      <div
        id="g_id_onload"
        data-client_id="975059814880-lfffftbpt7fdl14cevtve8sjvh015udc.apps.googleusercontent.com"
        data-context="signin"
        data-ux_mode="popup"
        data-callback="handleCredentialResponse"
        data-nonce=""
        data-itp_support="true"
      ></div>

      <div
        className="g_id_signin"
        data-type="standard"
        data-shape="rectangular"
        data-theme="filled_blue"
        data-text="signin_with"
        data-size="large"
        data-logo_alignment="left"
      ></div>
    </>
  )

  return (
    <>
      <Script src="https://accounts.google.com/gsi/client" async></Script>

      {/* Define callback function as string literal and set property on 'window' to avoid minimization of the
       *  callback function name.
       */}
      <Script id="google-auth-callback" strategy="afterInteractive">
        {`
          (function () {
            const setMessage = (text) => {
              const el = document.getElementById('login-message')
              if (!el) return
              el.textContent = text || ''
            }

            async function handleCredentialResponse(response) {
              if (!response || !response.credential) {
                setMessage('Missing login credential from Google.')
                return
              }

              try {
                const res = await fetch('/tokensignin', {
                  method: 'POST',
                  headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                  },
                  body: new URLSearchParams({
                    idtoken: response.credential,
                  }).toString(),
                })

                const data = await res.json().catch(() => ({}))
                if (!res.ok) {
                  setMessage('Error status code from server: ' + res.status)
                  return
                }

                if (data.redirect) {
                  window.location.href = '/'
                  return
                }

                setMessage(
                  data.message || 'Server could not authenticate you for some reason',
                )
              } catch (err) {
                setMessage('Could not reach the server. Please try again.')
              }
            }

            window.handleCredentialResponse = handleCredentialResponse
          })()
        `}
      </Script>

      <HydrationBoundary state={dehydrate(queryClient)}>
        <ContentWrapper size="sm" className="gap-6">
          <h1 className="text-lg">Log in to ADB</h1>

          {userJsx}

          <p
            id="login-message"
            className="text-sm text-red-500"
            aria-live="polite"
          ></p>

          {signInwithGoogleHtmlInvocation}
        </ContentWrapper>
      </HydrationBoundary>
    </>
  )
}
