import { useEffect, useState } from 'react'
import {
  GetStartupEnabled,
  SetStartupEnabled,
  IsStartupSupported,
} from '../../wailsjs/go/main/App'

const InfoIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="12" cy="12" r="10" />
    <line x1="12" y1="16" x2="12" y2="12" />
    <line x1="12" y1="8" x2="12.01" y2="8" />
  </svg>
)

const AlertCircleIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"
       strokeLinecap="round" strokeLinejoin="round">
    <circle cx="12" cy="12" r="10"/>
    <line x1="12" y1="8" x2="12" y2="12"/>
    <line x1="12" y1="16" x2="12.01" y2="16"/>
  </svg>
)

const XIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"
       strokeLinecap="round" strokeLinejoin="round">
    <line x1="18" y1="6" x2="6" y2="18"/>
    <line x1="6" y1="6" x2="18" y2="18"/>
  </svg>
)

function StartupPage() {
  const [isSupported, setIsSupported] = useState(false)
  const [isEnabled, setIsEnabled] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const loadState = async () => {
      try {
        const supported = await IsStartupSupported()
        setIsSupported(supported)

        if (supported) {
          const enabled = await GetStartupEnabled()
          setIsEnabled(enabled)
        }
      } catch (err) {
        console.error('Error loading startup state:', err)
        setError(`Failed to load startup settings: ${err}`)
      } finally {
        setIsLoading(false)
      }
    }

    loadState()
  }, [])

  const handleToggle = async () => {
    setError(null)
    const newValue = !isEnabled
    // Optimistically update UI
    setIsEnabled(newValue)
    try {
      await SetStartupEnabled(newValue)
    } catch (err) {
      console.error('Error toggling startup:', err)
      // Revert on error
      setIsEnabled(!newValue)
      setError(`Failed to ${newValue ? 'enable' : 'disable'} startup: ${err}`)
    }
  }

  const dismissError = () => {
    setError(null)
  }

  if (isLoading) {
    return (
      <div className="page">
        <div className="page-header">
          <h1 className="page-title">Startup</h1>
        </div>
        <div className="page-content">
          <div className="card">
            <div className="card-body">
              <div className="loading-state">Loading...</div>
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="page">
      <div className="page-header">
        <h1 className="page-title">Startup</h1>
        <p className="page-description">Configure application startup behavior</p>
      </div>

      <div className="page-content">
        {/* Error Banner */}
        {error && (
          <div className="error-banner">
            <div className="error-banner-content">
              <AlertCircleIcon />
              <span>{error}</span>
            </div>
            <button className="error-banner-dismiss" onClick={dismissError}>
              <XIcon />
            </button>
          </div>
        )}

        <div className="card">
          <div className="card-header">
            <h2 className="card-title">Launch on System Startup</h2>
          </div>
          <div className="card-body">
            <div className="setting-row">
              <div className="setting-info">
                <div className="setting-label">Run D2Tool when system starts</div>
                <div className="setting-description">
                  Automatically launch D2Tool when you log into your computer
                </div>
              </div>
              <label className={`toggle ${!isSupported ? 'disabled' : ''}`}>
                <input
                  type="checkbox"
                  checked={isEnabled}
                  onChange={handleToggle}
                  disabled={!isSupported}
                />
                <span className="toggle-slider"></span>
              </label>
            </div>

            {!isSupported && (
              <div className="alert alert-info">
                <InfoIcon />
                <span>Startup registration is only available on Windows</span>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default StartupPage
