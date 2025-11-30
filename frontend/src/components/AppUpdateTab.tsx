import { useEffect, useState } from 'react'
import { EventsOn } from '../wailsjs/runtime'
import {
  GetAppUpdateState,
  CheckForAppUpdate,
  DownloadAppUpdate,
  AppUpdateState,
} from '../wailsjs/go/main/App'

function AppUpdateTab() {
  const [state, setState] = useState<AppUpdateState>({
    currentVersion: '',
    latestVersion: '',
    lastCheckTime: 'Loading...',
    updateAvailable: false,
    isCheckingForUpdate: false,
    isDownloadingUpdate: false,
  })
  const [isChecking, setIsChecking] = useState(false)
  const [isDownloading, setIsDownloading] = useState(false)

  useEffect(() => {
    // Load initial state
    GetAppUpdateState().then(setState).catch(console.error)

    // Listen for update check events
    const offCheckStarted = EventsOn('appUpdateCheckStarted', () => {
      setIsChecking(true)
    })

    const offCheckFinished = EventsOn('appUpdateCheckFinished', (newState: AppUpdateState) => {
      setState(newState)
      setIsChecking(false)
    })

    // Listen for download events
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

  const isLoading = isChecking || isDownloading

  return (
    <div className="vbox">
      <p className="label label-bold">Last check time: {state.lastCheckTime}</p>

      <p className="label">Current version: {state.currentVersion}</p>

      {state.latestVersion && (
        <p className="label">Latest available version: {state.latestVersion}</p>
      )}

      <button
        className="button"
        onClick={handleCheckForUpdates}
        disabled={isLoading}
      >
        Check for updates
      </button>

      {state.updateAvailable && (
        <button
          className="button button-success"
          onClick={handleDownloadUpdate}
          disabled={isLoading}
        >
          Download update
        </button>
      )}

      {isLoading && (
        <div className="progress-bar">
          <div className="progress-bar-inner"></div>
        </div>
      )}
    </div>
  )
}

export default AppUpdateTab
