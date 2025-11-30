import { useEffect, useState } from 'react'
import { EventsOn } from '../wailsjs/runtime'
import {
  GetAppUpdateState,
  CheckForAppUpdate,
  DownloadAppUpdate,
  GetAutoUpdateEnabled,
  SetAutoUpdateEnabled,
  AppUpdateState,
} from '../wailsjs/go/main/App'

const SearchIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="11" cy="11" r="8" />
    <line x1="21" y1="21" x2="16.65" y2="16.65" />
  </svg>
)

const DownloadIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
    <polyline points="7 10 12 15 17 10" />
    <line x1="12" y1="15" x2="12" y2="3" />
  </svg>
)

const CheckCircleIcon = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
    <polyline points="22 4 12 14.01 9 11.01" />
  </svg>
)

const AlertCircleIcon = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="12" cy="12" r="10" />
    <line x1="12" y1="8" x2="12" y2="12" />
    <line x1="12" y1="16" x2="12.01" y2="16" />
  </svg>
)

const ClockIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="12" cy="12" r="10" />
    <polyline points="12 6 12 12 16 14" />
  </svg>
)

function UpdatesPage() {
  const [state, setState] = useState<AppUpdateState>({
    currentVersion: '',
    latestVersion: '',
    lastCheckTime: 'Loading...',
    updateAvailable: false,
    isCheckingForUpdate: false,
    isDownloadingUpdate: false,
    autoUpdateEnabled: true,
  })
  const [isChecking, setIsChecking] = useState(false)
  const [isDownloading, setIsDownloading] = useState(false)
  const [autoUpdateEnabled, setAutoUpdateEnabledState] = useState(true)

  useEffect(() => {
    GetAppUpdateState().then((s) => {
      setState(s)
      setAutoUpdateEnabledState(s.autoUpdateEnabled)
    }).catch(console.error)

    // Also get the current auto-update setting
    GetAutoUpdateEnabled().then(setAutoUpdateEnabledState).catch(console.error)

    const offCheckStarted = EventsOn('appUpdateCheckStarted', () => {
      setIsChecking(true)
    })

    const offCheckFinished = EventsOn('appUpdateCheckFinished', (newState: AppUpdateState) => {
      setState(newState)
      setAutoUpdateEnabledState(newState.autoUpdateEnabled)
      setIsChecking(false)
    })

    const offDownloadStarted = EventsOn('appUpdateDownloadStarted', () => {
      setIsDownloading(true)
    })

    const offDownloadFinished = EventsOn('appUpdateDownloadFinished', (result: { success: boolean; error: string }) => {
      setIsDownloading(false)
      if (result.success) {
        alert('Update downloaded successfully. Please restart the application.')
      } else {
        alert(`Error downloading update: ${result.error}`)
      }
    })

    return () => {
      offCheckStarted()
      offCheckFinished()
      offDownloadStarted()
      offDownloadFinished()
    }
  }, [])

  const handleCheckForUpdates = () => {
    CheckForAppUpdate().catch(console.error)
  }

  const handleDownloadUpdate = () => {
    DownloadAppUpdate().catch(console.error)
  }

  const handleAutoUpdateToggle = async () => {
    const newValue = !autoUpdateEnabled
    setAutoUpdateEnabledState(newValue)
    try {
      await SetAutoUpdateEnabled(newValue)
    } catch (error) {
      console.error('Error toggling auto-update:', error)
      // Revert on error
      setAutoUpdateEnabledState(!newValue)
    }
  }

  const isLoading = isChecking || isDownloading

  return (
    <div className="page">
      <div className="page-header">
        <h1 className="page-title">Updates</h1>
        <p className="page-description">Check for and install application updates</p>
      </div>

      <div className="page-content">
        {/* Auto-Update Settings Card */}
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">Update Settings</h2>
          </div>
          <div className="card-body">
            <div className="setting-row">
              <div className="setting-info">
                <div className="setting-label">Automatic Updates</div>
                <div className="setting-description">
                  Automatically check for updates when D2Tool starts
                </div>
              </div>
              <label className="toggle">
                <input
                  type="checkbox"
                  checked={autoUpdateEnabled}
                  onChange={handleAutoUpdateToggle}
                />
                <span className="toggle-slider"></span>
              </label>
            </div>
          </div>
        </div>

        {/* Version Info Card */}
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">Version Information</h2>
          </div>
          <div className="card-body">
            <div className="version-grid">
              <div className="version-item">
                <div className="version-label">Current Version</div>
                <div className="version-value">{state.currentVersion || 'Unknown'}</div>
              </div>
              {state.latestVersion && (
                <div className="version-item">
                  <div className="version-label">Latest Version</div>
                  <div className="version-value">{state.latestVersion}</div>
                </div>
              )}
            </div>

            <div className="status-row mt-16">
              <div className="status-info">
                <div className="status-label">
                  <ClockIcon />
                  <span>Last Checked</span>
                </div>
                <div className="status-value">{state.lastCheckTime}</div>
              </div>
            </div>
          </div>
        </div>

        {/* Update Status Card */}
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">Update Status</h2>
          </div>
          <div className="card-body">
            {state.updateAvailable ? (
              <div className="update-available">
                <div className="update-badge update-badge-warning">
                  <AlertCircleIcon />
                  <span>Update Available</span>
                </div>
                <p className="update-message">
                  A new version ({state.latestVersion}) is available. Download now to get the latest features and improvements.
                </p>
              </div>
            ) : (
              <div className="update-available">
                <div className="update-badge update-badge-success">
                  <CheckCircleIcon />
                  <span>Up to Date</span>
                </div>
                <p className="update-message">
                  You're running the latest version of D2Tool.
                </p>
              </div>
            )}

            <div className="button-group mt-16">
              <button
                className="btn btn-secondary"
                onClick={handleCheckForUpdates}
                disabled={isLoading}
              >
                <SearchIcon />
                <span>{isChecking ? 'Checking...' : 'Check for Updates'}</span>
              </button>

              {state.updateAvailable && (
                <button
                  className="btn btn-primary"
                  onClick={handleDownloadUpdate}
                  disabled={isLoading}
                >
                  <DownloadIcon />
                  <span>{isDownloading ? 'Downloading...' : 'Download Update'}</span>
                </button>
              )}
            </div>

            {isLoading && (
              <div className="progress-bar mt-16">
                <div className="progress-bar-inner"></div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default UpdatesPage
