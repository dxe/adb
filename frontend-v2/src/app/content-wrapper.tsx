import { cn } from '@/lib/utils'
import { ReactNode } from 'react'

const contentWrapperClass = {
  sm: 'lg:max-w-lg',
  md: 'lg:max-w-screen-md',
  lg: 'lg:max-w-screen-lg',
  xl: 'lg:max-w-screen-xl',
  '2xl': 'lg:max-w-screen-2xl',
}

/**
 * ContentWrapper is the white box of variable width that contains all of the
 * page elements. It is equivalent to the "body wrapper" element in the Vue app.
 */
export const ContentWrapper = (props: {
  size: keyof typeof contentWrapperClass
  /**
   * Whether to make the content wrapper take the full width of the viewport.
   * Be sure to keep in sync with `isFullScreenPage` in
   * site-background-controller.tsx
   */
  fullScreen?: boolean
  className?: string
  children: ReactNode
}) => {
  return (
    <div
      className={cn(
        'bg-white w-full py-6 px-4 md:px-10 flex flex-col',
        props.fullScreen
          ? 'lg:max-w-none lg:mt-0 lg:mx-0 lg:rounded-none shadow-none bg-opacity-100'
          : 'lg:rounded-md shadow-2xl backdrop-blur-md bg-opacity-95 lg:mt-6 lg:mx-auto',
        !props.fullScreen && contentWrapperClass[props.size],
        props.className,
      )}
    >
      {props.children}
    </div>
  )
}
