import { useEffect, useState } from 'react'
import { EventsOn } from '../../wailsjs/runtime'
import {
  GetSteamConfig,
  SetSteamPath,
  SetAutoEnableNewAccounts,
  GetSteamAccounts,
  SetSteamAccountEnabled,
  OpenDirectoryDialog,
  IsSteamPathValid,
  RescanSteamAccounts,
} from '../../wailsjs/go/main/App'
import { config, steam } from '../../wailsjs/go/models'
import { EventSteamAccountsChanged, EventSteamPathChanged } from '../events'
import AccountCard from '../components/AccountCard'
import { AlertCircleIcon, FolderIcon, RefreshIcon } from '../components/Icons'
import { useGridAutoUpdate } from '../components/GridAutoUpdateProvider'

function SteamPage() {
  const [steamConfig, setSteamConfig] = useState<config.SteamConfig | null>(null)
  const [accounts, setAccounts] = useState<steam.SteamAccountView[]>([])
  const [pathValid, setPathValid] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isScanning, setIsScanning] = useState(false)

  const { scheduleGridUpdate } = useGridAutoUpdate()

  const refresh = () => {
    Promise.all([
      GetSteamConfig(),
      GetSteamAccounts(),
      IsSteamPathValid(),
    ]).then(([cfg, accs, valid]) => {
      setSteamConfig(cfg)
      setAccounts(accs)
      setPathValid(valid)
    }).catch(console.error)
  }

  useEffect(() => {
    refresh()

    const offAccountsChanged = EventsOn(EventSteamAccountsChanged, refresh)
    const offPathChanged = EventsOn(EventSteamPathChanged, refresh)

    return () => {
      offAccountsChanged()
      offPathChanged()
    }
  }, [])

  const handleBrowsePath = async () => {
    try {
      const path = await OpenDirectoryDialog()
      if (path) {
        setError(null)
        await SetSteamPath(path)
      }
    } catch (err) {
      setError(`Failed to set Steam path: ${err}`)
    }
  }

  const handleToggleAutoEnable = async (enabled: boolean) => {
    try {
      await SetAutoEnableNewAccounts(enabled)
      GetSteamConfig().then(setSteamConfig).catch(console.error)
    } catch (err) {
      console.error('Error toggling auto-enable:', err)
    }
  }

  const handleToggleAccountEnabled = async (steamId64: string, enabled: boolean) => {
    try {
      await SetSteamAccountEnabled(steamId64, enabled)
      setAccounts(prev => prev.map(a =>
        a.steamId64 === steamId64 ? { ...a, enabled } : a
      ))
      scheduleGridUpdate()
    } catch (err) {
      console.error('Error toggling account:', err)
    }
  }

  const handleRescan = async () => {
    setIsScanning(true)
    setError(null)
    try {
      await RescanSteamAccounts()
    } catch (err) {
      setError(`Failed to rescan accounts: ${err}`)
    } finally {
      setIsScanning(false)
    }
  }

  return (
    <div className="page">
      <div className="page-header">
        <div className="page-header-text">
          <h1 className="page-title">Steam</h1>
          <p className="page-description">Manage Steam integration and account settings</p>
        </div>
      </div>

      <div className="page-content">
        {error && (
          <div className="alert alert-error">
            <AlertCircleIcon />
            <span>{error}</span>
          </div>
        )}

        {/* Steam Directory Card */}
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">Steam Directory</h2>
          </div>
          <div className="card-body">
            {!pathValid && (
              <div className="alert alert-warning">
                <AlertCircleIcon />
                <span>Steam directory not found. Please set the path manually.</span>
              </div>
            )}
            <div className="setting-row">
              <div className="setting-info">
                <div className="setting-label">Steam Path</div>
                <div className="setting-description">
                  {steamConfig?.steamPath || 'Not set'}
                </div>
              </div>
              <button className="btn btn-secondary btn-sm" onClick={handleBrowsePath}>
                <FolderIcon />
                <span>Browse</span>
              </button>
            </div>
          </div>
        </div>

        {/* Steam Accounts Card */}
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">Steam Accounts</h2>
            <button
              className="btn btn-secondary btn-sm"
              onClick={handleRescan}
              disabled={isScanning}
            >
              <RefreshIcon />
              <span>{isScanning ? 'Scanning...' : 'Rescan'}</span>
            </button>
          </div>
          <div className="card-body">
            {accounts.length === 0 ? (
              <div className="empty-state">
                <p>No Steam accounts found</p>
                <p className="empty-state-hint">Make sure Steam path is correct and you have played Dota 2 on at least one account</p>
              </div>
            ) : (
              <div className="file-list">
                {accounts.map((account) => (
                  <AccountCard
                    key={account.steamId64}
                    account={account}
                    toggle={{
                      checked: account.enabled,
                      onChange: (enabled) => handleToggleAccountEnabled(account.steamId64, enabled),
                    }}
                  />
                ))}
              </div>
            )}
          </div>
        </div>

        {/* New Account Discovery Card */}
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">New Account Discovery</h2>
          </div>
          <div className="card-body">
            <div className="setting-row">
              <div className="setting-info">
                <div className="setting-label">Auto-enable new accounts</div>
                <div className="setting-description">
                  When a new Steam account is discovered, automatically enable it for hero layout updates
                </div>
              </div>
              <label className="toggle">
                <input
                  type="checkbox"
                  checked={steamConfig?.autoEnableNewAccounts ?? true}
                  onChange={(e) => handleToggleAutoEnable(e.target.checked)}
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

export default SteamPage
