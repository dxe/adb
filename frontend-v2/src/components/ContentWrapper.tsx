import { cn } from "@/util/cn";
import { ReactNode } from "react";

const contentWrapperClass = {
  sm: "lg:max-w-screen-sm",
  md: "lg:max-w-screen-md",
  lg: "lg:max-w-screen-lg",
  xl: "lg:max-w-screen-xl",
  "2xl": "lg:max-w-screen-2xl",
};

export const ContentWrapper = (props: {
  size: keyof typeof contentWrapperClass;
  children: ReactNode;
}) => {
  return (
    <div
      className={cn(
        "bg-white w-full rounded-md py-6 px-10 shadow-2xl backdrop-blur-md bg-opacity-95 lg:mt-6 lg:mx-auto",
        contentWrapperClass[props.size]
      )}
    >
      {props.children}
    </div>
  );
};
