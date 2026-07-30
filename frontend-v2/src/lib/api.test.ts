import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

// ApiClient uses a "/" prefix in the browser, producing root-relative request
// URLs that a real browser resolves against the document. Under jsdom the
// global Request comes from node's undici, which rejects relative URLs, so
// shim it to resolve them against a base origin for the duration of the test.
const RealRequest = globalThis.Request
class BaseResolvingRequest extends RealRequest {
  constructor(input: RequestInfo | URL, init?: RequestInit) {
    if (typeof input === 'string' && input.startsWith('/')) {
      input = `http://localhost${input}`
    }
    super(input, init)
  }
}

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'content-type': 'application/json' },
  })
}

type FetchHandler = (req: Request) => Response | Promise<Response>

function stubFetch(handler: FetchHandler) {
  vi.stubGlobal(
    'fetch',
    vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      const req = input instanceof Request ? input : new Request(input, init)
      return handler(req)
    }),
  )
}

async function loadApi() {
  vi.resetModules()
  return import('./api')
}

beforeEach(() => {
  vi.stubGlobal('Request', BaseResolvingRequest)
})

afterEach(() => {
  vi.unstubAllGlobals()
})

// These tests exercise the CSRF invalidate-and-retry path in ApiClient. The
// token is cached at module scope, so each test loads a fresh copy of the
// module (vi.resetModules) to start from an empty cache, and drives the real ky
// client by stubbing globalThis.fetch.
describe('ApiClient CSRF handling', () => {
  it('refetches the token and retries once when a stale token is rejected', async () => {
    const tokens = ['stale-token', 'fresh-token']
    let tokenIdx = 0
    const sentTokens: (string | null)[] = []
    let postCalls = 0

    stubFetch(async (req) => {
      const { pathname } = new URL(req.url)
      if (pathname.endsWith('/api/csrf-token')) {
        return jsonResponse({
          status: 'success',
          csrfToken: tokens[tokenIdx++],
        })
      }
      if (pathname.endsWith('/api/admin/send-test-email')) {
        postCalls++
        sentTokens.push(req.headers.get('X-CSRF-Token'))
        // First attempt carries the stale token; the server rejects it the way
        // gorilla/csrf does once the underlying cookie has rotated.
        if (postCalls === 1) {
          return new Response('Forbidden - CSRF token invalid', { status: 403 })
        }
        return jsonResponse({ status: 'success' })
      }
      throw new Error(`unexpected request: ${req.url}`)
    })

    const { apiClient } = await loadApi()
    const result = await apiClient.sendTestEmail('test@example.com')

    expect(result).toEqual({ status: 'success' })
    // The mutation was attempted twice, with the refreshed token on the retry.
    expect(postCalls).toBe(2)
    expect(sentTokens).toEqual(['stale-token', 'fresh-token'])
  })

  it('does not retry a plain authorization 403', async () => {
    let postCalls = 0

    stubFetch(async (req) => {
      const { pathname } = new URL(req.url)
      if (pathname.endsWith('/api/csrf-token')) {
        return jsonResponse({ status: 'success', csrfToken: 'token' })
      }
      if (pathname.endsWith('/api/admin/send-test-email')) {
        postCalls++
        // Authorization denial: bare "Forbidden", no CSRF failure reason.
        return new Response('Forbidden', { status: 403 })
      }
      throw new Error(`unexpected request: ${req.url}`)
    })

    const { apiClient } = await loadApi()
    await expect(apiClient.sendTestEmail('test@example.com')).rejects.toThrow()
    // No refetch/retry: the request was attempted exactly once.
    expect(postCalls).toBe(1)
  })
})
