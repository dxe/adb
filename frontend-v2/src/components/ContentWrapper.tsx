import { cn } from '@/lib/utils'
import { ReactNode } from 'react'

const contentWrapperClass = {
  sm: 'lg:max-w-screen-sm',
  md: 'lg:max-w-screen-md',
  lg: 'lg:max-w-screen-lg',
  xl: 'lg:max-w-screen-xl',
  '2xl': 'lg:max-w-screen-2xl',
}

/** ContentWrapper is equivalent to the "body wrapper" element that was used in Vue.
 *  It is the white box of varialbe width that contains all of the page elements.
 */
export const ContentWrapper = (props: {
  size: keyof typeof contentWrapperClass
  className?: string
  children: ReactNode
}) => {
  return (
    <div
      className={cn(
        'bg-white w-full lg:rounded-md py-6 px-10 shadow-2xl backdrop-blur-md bg-opacity-95 lg:mt-6 lg:mx-auto flex flex-col',
        contentWrapperClass[props.size],
        props.className,
      )}
    >
      {props.children}
    </div>
  )
}
