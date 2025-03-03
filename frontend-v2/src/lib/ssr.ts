import { dehydrate, QueryClient } from '@tanstack/react-query'
import { API_PATH, ApiClient } from '@/lib/api'
import { GetServerSidePropsContext, InferGetServerSidePropsType } from 'next'

/** Loads the static resource hash & session data during SSR.
 *  On any page that you want to pre-fetch these things,
 *  this function should be imported and then exported
 *  as `getServerSideProps`, i.e.
 *  `export const getServerSideProps = getDefaultServerSideProps`.
 *  When using this, be sure that you also wrap the page in the
 *  HydrationBoundary so that it gets hydrated properly.
 */
export const getDefaultServerSideProps = async (
  context: GetServerSidePropsContext,
) => {
  const cookies = context.req.headers.cookie
  const ssrApiClient = new ApiClient(cookies)
  const queryClient = new QueryClient()
  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: [API_PATH.STATIC_RESOURCE_HASH],
      queryFn: ssrApiClient.getStaticResourceHash,
    }),
    queryClient.prefetchQuery({
      queryKey: [API_PATH.USER_ME],
      queryFn: ssrApiClient.getAuthedUser,
    }),
  ])
  return {
    props: {
      dehydratedState: dehydrate(queryClient),
    },
  }
}

export type DefaultPageProps = InferGetServerSidePropsType<
  typeof getDefaultServerSideProps
>
