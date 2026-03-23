export function formatLastUpdate(timestampMillis: number) {
  if (timestampMillis === 0) return 'Never'
  return new Date(timestampMillis).toLocaleString()
}
