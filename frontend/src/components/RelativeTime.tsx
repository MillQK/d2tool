import { useEffect, useState } from 'react'
import { formatRelativeTime, formatFullDate, getRefreshInterval } from '../utils/format'

interface RelativeTimeProps {
  timestampMillis: number
  prefix?: string
  neverText?: string
  className?: string
}

function RelativeTime({ timestampMillis, prefix, neverText = 'Never updated', className = 'file-status' }: RelativeTimeProps) {
  const [, setTick] = useState(0)

  useEffect(() => {
    const interval = getRefreshInterval(timestampMillis)
    if (interval === 0) return

    const id = setInterval(() => setTick(t => t + 1), interval)
    return () => clearInterval(id)
  }, [timestampMillis])

  if (timestampMillis === 0) {
    return <span className={className}>{neverText}</span>
  }

  const relative = formatRelativeTime(timestampMillis)
  const full = formatFullDate(timestampMillis)
  const text = prefix ? `${prefix}${relative}` : relative

  return (
    <span className={className} title={full}>
      {text}
    </span>
  )
}

export default RelativeTime
