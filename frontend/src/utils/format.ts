export function formatFullDate(timestampMillis: number): string {
  if (timestampMillis === 0) return 'Never'
  return new Date(timestampMillis).toLocaleString('en-GB', {
    hour12: false,
  })
}

export function formatRelativeTime(timestampMillis: number): string {
  if (timestampMillis === 0) return 'Never'

  const seconds = Math.floor((Date.now() - timestampMillis) / 1000)

  if (seconds < 60) return 'Just now'

  const minutes = Math.floor(seconds / 60)
  if (minutes === 1) return '1 min ago'
  if (minutes < 60) return `${minutes} mins ago`

  const hours = Math.floor(minutes / 60)
  if (hours === 1) return '1 hour ago'
  if (hours < 24) return `${hours} hours ago`

  const days = Math.floor(hours / 24)
  if (days === 1) return 'Yesterday'
  if (days < 7) return `${days} days ago`

  return formatFullDate(timestampMillis)
}

// Returns an appropriate refresh interval in ms based on how old the timestamp is.
// Recent times need frequent updates, older times can refresh less often.
export function getRefreshInterval(timestampMillis: number): number {
  if (timestampMillis === 0) return 0

  const seconds = Math.floor((Date.now() - timestampMillis) / 1000)

  if (seconds < 60) return 10_000       // "Just now" — refresh every 10s
  if (seconds < 3600) return 30_000      // "N mins ago" — refresh every 30s
  if (seconds < 86400) return 60_000     // "N hours ago" — refresh every 1min
  if (seconds < 604800) return 300_000   // "N days ago" — refresh every 5min
  return 0                               // Full date — no refresh needed
}
