import { useEffect, useState } from 'react'
import {
  GetStartupEnabled,
  SetStartupEnabled,
  IsStartupSupported,
} from '../wailsjs/go/main/App'

function StartupTab() {
  const [isSupported, setIsSupported] = useState(false)
  const [isEnabled, setIsEnabled] = useState(false)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const loadState = async () => {
      try {
        const supported = await IsStartupSupported()
        setIsSupported(supported)

        if (supported) {
          const enabled = await GetStartupEnabled()
          setIsEnabled(enabled)
        }
      } catch (error) {
        console.error('Error loading startup state:', error)
      } finally {
        setIsLoading(false)
      }
    }

    loadState()
  }, [])

  const handleToggle = async () => {
    try {
      const newValue = !isEnabled
      await SetStartupEnabled(newValue)
      setIsEnabled(newValue)
    } catch (error) {
      console.error('Error toggling startup:', error)
    }
  }

  if (isLoading) {
    return <div className="vbox"><p className="label">Loading...</p></div>
  }

  return (
    <div className="vbox">
      <label className={`checkbox-container ${!isSupported ? 'disabled' : ''}`}>
        <input
          type="checkbox"
          checked={isEnabled}
          onChange={handleToggle}
          disabled={!isSupported}
        />
        <span>Run on startup</span>
      </label>

      {!isSupported && (
        <p className="label" style={{ color: '#8b9cac', fontStyle: 'italic' }}>
          Startup registration is only available on Windows.
        </p>
      )}
    </div>
  )
}

export default StartupTab
