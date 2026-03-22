import {useState} from 'react'
import {
    CheckForAppUpdate,
    DownloadAppUpdate,
} from '../../wailsjs/go/main/App'
import {Quit} from '../../wailsjs/runtime'
import {main} from "../../wailsjs/go/models.ts";
import { XIcon, AlertCircleIcon, DownloadIcon, SearchIcon, CheckCircleIcon, ClockIcon } from '../components/Icons'

interface DownloadResult {
    success: boolean
    message: string
}

interface UpdatesPageProps {
    state: main.AppUpdateState | null
    onStateChange: () => void
}

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
                            <AlertCircleIcon size={20}/>
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
                                    <AlertCircleIcon size={20}/>
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
                                        {downloadResult.success ? <CheckCircleIcon/> : <AlertCircleIcon size={20}/>}
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
