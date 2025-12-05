import {useState} from 'react'
import {
    CheckForAppUpdate,
    DownloadAppUpdate,
} from '../../wailsjs/go/main/App'
import {Quit} from '../../wailsjs/runtime'
import {main} from "../../wailsjs/go/models.ts";

interface DownloadResult {
    success: boolean
    message: string
}

interface UpdatesPageProps {
    state: main.AppUpdateState | null
    onStateChange: () => void
}

const SearchIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"
         strokeLinecap="round" strokeLinejoin="round">
        <circle cx="11" cy="11" r="8"/>
        <line x1="21" y1="21" x2="16.65" y2="16.65"/>
    </svg>
)

const DownloadIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"
         strokeLinecap="round" strokeLinejoin="round">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
        <polyline points="7 10 12 15 17 10"/>
        <line x1="12" y1="15" x2="12" y2="3"/>
    </svg>
)

const CheckCircleIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"
         strokeLinecap="round" strokeLinejoin="round">
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
        <polyline points="22 4 12 14.01 9 11.01"/>
    </svg>
)

const AlertCircleIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"
         strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10"/>
        <line x1="12" y1="8" x2="12" y2="12"/>
        <line x1="12" y1="16" x2="12.01" y2="16"/>
    </svg>
)

const ClockIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"
         strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10"/>
        <polyline points="12 6 12 12 16 14"/>
    </svg>
)

const XIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"
         strokeLinecap="round" strokeLinejoin="round">
        <line x1="18" y1="6" x2="6" y2="18"/>
        <line x1="6" y1="6" x2="18" y2="18"/>
    </svg>
)

function UpdatesPage({ state, onStateChange }: UpdatesPageProps) {
    const [isChecking, setIsChecking] = useState(false)
    const [isDownloading, setIsDownloading] = useState(false)
    const [downloadResult, setDownloadResult] = useState<DownloadResult | null>(null)
    const [checkError, setCheckError] = useState<string | null>(null)

    const handleCheckForUpdates = async () => {
        setIsChecking(true)
        setCheckError(null)
        try {
            await CheckForAppUpdate()
            onStateChange()
        } catch (error) {
            console.error('Error checking for updates:', error)
            setCheckError(`Failed to check for updates: ${error}`)
        } finally {
            setIsChecking(false)
        }
    }

    const dismissCheckError = () => {
        setCheckError(null)
    }

    const handleDownloadUpdate = async () => {
        setIsDownloading(true)
        setDownloadResult(null)
        try {
            await DownloadAppUpdate()
            setDownloadResult({
                success: true,
                message: 'Update downloaded successfully. Please restart the application to apply the update.'
            })
        } catch (error) {
            console.error('Error downloading update:', error)
            setDownloadResult({
                success: false,
                message: `Error downloading update: ${error}. Please try again.`
            })
        } finally {
            setIsDownloading(false)
        }
    }

    const handleQuit = () => {
        Quit()
    }

    const dismissResult = () => {
        setDownloadResult(null)
    }

    const isLoading = isChecking || isDownloading

    return (
        <div className="page">
            <div className="page-header">
                <h1 className="page-title">Updates</h1>
                <p className="page-description">Check for and install application updates</p>
            </div>

            <div className="page-content">
                {/* Error Banner */}
                {checkError && (
                    <div className="error-banner">
                        <div className="error-banner-content">
                            <AlertCircleIcon/>
                            <span>{checkError}</span>
                        </div>
                        <button className="error-banner-dismiss" onClick={dismissCheckError}>
                            <XIcon/>
                        </button>
                    </div>
                )}

                {/* Version Info Card */}
                <div className="card">
                    <div className="card-header">
                        <h2 className="card-title">Version Information</h2>
                    </div>
                    <div className="card-body">
                        <div className="version-grid">
                            <div className="version-item">
                                <div className="version-label">Current Version</div>
                                <div className="version-value">{state?.currentVersion || 'Unknown'}</div>
                            </div>
                            {state?.latestVersion && (
                                <div className="version-item">
                                    <div className="version-label">Latest Version</div>
                                    <div className="version-value">{state.latestVersion}</div>
                                </div>
                            )}
                        </div>

                        <div className="status-row mt-16">
                            <div className="status-info">
                                <div className="status-label">
                                    <ClockIcon/>
                                    <span>Last Checked</span>
                                </div>
                                <div className="status-value">{state?.lastCheckTime || 'Loading...'}</div>
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
                        {state?.updateAvailable ? (
                            <div className="update-available">
                                <div className="update-badge update-badge-warning">
                                    <AlertCircleIcon/>
                                    <span>Update Available</span>
                                </div>
                                <p className="update-message">
                                    A new version ({state.latestVersion}) is available. Download now to get the latest
                                    features and improvements.
                                </p>
                            </div>
                        ) : (
                            <div className="update-available">
                                <div className="update-badge update-badge-success">
                                    <CheckCircleIcon/>
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
                                <SearchIcon/>
                                <span>{isChecking ? 'Checking...' : 'Check for Updates'}</span>
                            </button>

                            {state?.updateAvailable && (
                                <button
                                    className="btn btn-primary"
                                    onClick={handleDownloadUpdate}
                                    disabled={isLoading}
                                >
                                    <DownloadIcon/>
                                    <span>{isDownloading ? 'Downloading...' : 'Download Update'}</span>
                                </button>
                            )}
                        </div>

                        {isLoading && (
                            <div className="progress-bar mt-16">
                                <div className="progress-bar-inner"></div>
                            </div>
                        )}

                        {downloadResult && (
                            <div className={`download-result mt-16 ${downloadResult.success ? 'download-result-success' : 'download-result-error'}`}>
                                <div className="download-result-content">
                                    <div className="download-result-icon">
                                        {downloadResult.success ? <CheckCircleIcon/> : <AlertCircleIcon/>}
                                    </div>
                                    <div className="download-result-text">
                                        <p>{downloadResult.message}</p>
                                        {downloadResult.success && (
                                            <button className="btn-quit" onClick={handleQuit}>
                                                Quit the application
                                            </button>
                                        )}
                                    </div>
                                </div>
                                <button className="download-result-dismiss" onClick={dismissResult}>
                                    <XIcon/>
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    )
}

export default UpdatesPage
