# Components/UI

Components in this directory are generated via [shadcn](https://ui.shadcn.com/docs).

## Customizations

When upgrading shadcn components, preserve the following customizations:

### select.tsx

SelectTrigger includes custom border styling for improved visual feedback:
- `hover:border-gray-400` - Border color on hover
- `focus:border-primary` - Border color when focused
- `focus:hover:border-primary` - Maintains primary border when both focused and hovered
