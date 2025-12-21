import { useEffect, useState } from 'react'
import { GetD2PTConfig, SetD2PTPeriod } from '../../wailsjs/go/main/App'
import { config } from '../../wailsjs/go/models'

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

// Period options for D2PT provider
const periodOptions = [
  { value: '8', label: 'Last 8 days' },
  { value: 'patch', label: 'Current patch' },
]

function ProvidersPage() {
  const [d2ptConfig, setD2ptConfig] = useState<config.D2PTConfig | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const loadConfig = async () => {
      try {
        const cfg = await GetD2PTConfig()
        setD2ptConfig(cfg)
      } catch (err) {
        console.error('Error loading D2PT config:', err)
        setError(`Failed to load provider settings: ${err}`)
      } finally {
        setIsLoading(false)
      }
    }

    loadConfig()
  }, [])

  const handlePeriodChange = async (newPeriod: string) => {
    setError(null)
    try {
      await SetD2PTPeriod(newPeriod)
      if (d2ptConfig) {
        setD2ptConfig({ ...d2ptConfig, period: newPeriod })
      }
    } catch (err) {
      console.error('Error setting D2PT period:', err)
      setError(`Failed to update period: ${err}`)
    }
  }

  const dismissError = () => {
    setError(null)
  }

  if (isLoading) {
    return (
      <div className="page">
        <div className="page-header">
          <h1 className="page-title">Providers</h1>
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
        <h1 className="page-title">Providers</h1>
        <p className="page-description">Configure data provider settings</p>
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

        {/* D2PT Provider Card */}
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">Dota2ProTracker</h2>
          </div>
          <div className="card-body">
            <div className="setting-row">
              <div className="setting-info">
                <div className="setting-label">Period</div>
                <div className="setting-description">
                  Time period for hero statistics data
                </div>
              </div>
              <select
                className="select"
                value={d2ptConfig?.period || '8'}
                onChange={(e) => handlePeriodChange(e.target.value)}
              >
                {periodOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default ProvidersPage
