import type { ChangeEvent } from 'react'
import { SetAutoUpdateEnabled } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { AlertCircleIcon, CheckCircleIcon } from '../components/Icons'

interface UpdatesPageProps {
  state: main.AppUpdateState | null
}

function UpdatesPage({ state }: UpdatesPageProps) {
  const handleAutoUpdateChange = (event: ChangeEvent<HTMLInputElement>) => {
    SetAutoUpdateEnabled(event.target.checked).catch((error) => {
      console.error('Error changing automatic update setting:', error)
    })
  }

  return (
    <div className="page">
      <div className="page-header">
        <div className="page-header-text">
          <h1 className="page-title">Updates</h1>
          <p className="page-description">Keep D2Tool up to date automatically</p>
        </div>
      </div>

      <div className="page-content">
        <div className="card">
          <div className="card-header"><h2 className="card-title">Update Settings</h2></div>
          <div className="card-body">
            <div className="setting-row">
              <div className="setting-info">
                <div className="setting-label">Automatic Updates</div>
                <div className="setting-description">Download new versions automatically and use them on the next start.</div>
              </div>
              <label className={`toggle ${state === null ? 'disabled' : ''}`}>
                <input
                  type="checkbox"
                  checked={state?.autoUpdateEnabled ?? true}
                  disabled={state === null}
                  onChange={handleAutoUpdateChange}
                />
                <span className="toggle-slider" />
              </label>
            </div>
          </div>
        </div>

        <div className="card">
          <div className="card-header"><h2 className="card-title">Version Information</h2></div>
          <div className="card-body">
            <div className="version-grid">
              <div className="version-item">
                <div className="version-label">Current Version</div>
                <div className="version-value">{state?.currentVersion || 'Unknown'}</div>
              </div>
              <div className="version-item">
                <div className="version-label">Latest Version</div>
                <div className="version-value">{state?.latestVersion || 'Unknown'}</div>
              </div>
            </div>
          </div>
        </div>

        {state?.status === 'updating' && (
          <div className="card">
            <div className="card-body">
              <div className="update-badge update-badge-warning"><span>Updating…</span></div>
              <div className="progress-bar"><div className="progress-bar-inner" /></div>
            </div>
          </div>
        )}

        {state?.status === 'ready' && (
          <div className="card">
            <div className="card-body update-available">
              <div className="update-badge update-badge-success"><CheckCircleIcon /><span>Will be updated on next start</span></div>
            </div>
          </div>
        )}

        {state?.status === 'error' && (
          <div className="error-banner">
            <div className="error-banner-content">
              <AlertCircleIcon size={20} />
              <span>{state.errorMessage || 'Automatic update failed. It will be retried automatically.'}</span>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default UpdatesPage
