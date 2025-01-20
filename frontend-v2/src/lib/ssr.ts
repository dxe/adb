import { dehydrate, QueryClient, QueryFunction } from '@tanstack/react-query'
import { API_PATH, getAuthedUser, getStaticResourceHash } from '@/lib/api'
import { InferGetServerSidePropsType } from 'next'

/** Loads the static resource hash & session data during SSR.
 *  On any page that you want to pre-fetch these things,
 *  this function should be imported and then exported
 *  as `getServerSideProps`, i.e.
 *  `export const getServerSideProps = getDefaultServerSideProps`.
 *  When using this, be sure that you also wrap the page in the
 *  HydrationBoundary so that it gets hydrated properly.
 */
export const getDefaultServerSideProps = async (
  /** If you want to prefetch additional queries than just the
   *  defaults that we fetch for every page, you can add more here.
   */
  prefetchAdditionalQueries?: { queryKey: string[]; queryFn: QueryFunction }[],
) => {
  const queryClient = new QueryClient()
  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: [API_PATH.STATIC_RESOURCE_HASH],
      queryFn: getStaticResourceHash,
    }),
    queryClient.prefetchQuery({
      queryKey: [API_PATH.USER_ME],
      queryFn: getAuthedUser,
    }),
    ...(prefetchAdditionalQueries?.length
      ? prefetchAdditionalQueries.map((q) =>
          queryClient.prefetchQuery({ ...q }),
        )
      : []),
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
