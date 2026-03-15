# Shared compatibility notes

Files in `shared/` may be imported by both:

- `frontend-v2` (Next/modern TS)
- `frontend` (legacy Vue 2 + webpack 4 + TypeScript 3.6)

If a file here is used by the Vue app, keep its JavaScript/TypeScript syntax compatible with that older toolchain.

In practice:

- avoid newer syntax like `?.` and `??`
- avoid inline type-only imports like `import { type X } from '...'`
- prefer plain ESM and simple TypeScript types

If you use newer syntax in `shared/`, `frontend-v2` may work while the legacy Vue build breaks.
