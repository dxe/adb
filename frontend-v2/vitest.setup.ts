import '@testing-library/jest-dom/vitest'

// jsdom doesn't implement ResizeObserver, which Radix's ScrollArea uses on mount.
if (typeof globalThis.ResizeObserver === 'undefined') {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
}
