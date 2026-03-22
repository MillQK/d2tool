import { useEffect, useState } from 'react'
import {
  GetStartupEnabled,
  SetStartupEnabled,
  IsStartupSupported,
} from '../../wailsjs/go/main/App'
import { AlertCircleIcon, InfoIcon, XIcon } from '../components/Icons'

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
            {!isSupported && (
                <div className="alert alert-info">
                  <InfoIcon />
                  <span>Startup registration is only available on Windows</span>
                </div>
            )}
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
          </div>
        </div>
      </div>
    </div>
  )
}

export default StartupPage
