import { useState, KeyboardEvent } from 'react'

export function useSuggestions(getSuggestions: (value: string) => string[]) {
  const [suggestions, setSuggestions] = useState<string[]>([])
  const [selectedIndex, setSelectedIndex] = useState(-1)

  const onInputChange = (value: string) => {
    setSuggestions(getSuggestions(value))
    setSelectedIndex(-1)
  }

  const onSelect = () => {
    setSuggestions([])
    setSelectedIndex(-1)
  }

  // Handles ArrowUp, ArrowDown, and Escape. Callers handle Enter/Tab themselves.
  const onKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setSelectedIndex((i) => (i === suggestions.length - 1 ? 0 : i + 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setSelectedIndex((i) => (i <= 0 ? suggestions.length - 1 : i - 1))
    } else if (e.key === 'Escape') {
      setSuggestions([])
      setSelectedIndex(-1)
    }
  }

  return { suggestions, selectedIndex, onInputChange, onSelect, onKeyDown }
}
