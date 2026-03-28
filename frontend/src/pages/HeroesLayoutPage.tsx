import { useEffect, useRef, useState } from 'react'
import { EventsOn } from '../../wailsjs/runtime'
import {
  UpdateHeroesLayout,
  GetHeroesLayoutFiles,
  AddHeroesLayoutFile,
  RemoveHeroesLayoutFile,
  SetHeroesLayoutFileEnabled,
  OpenFileDialog,
  GetPositions,
  SetPositions,
  SetPositionEnabled,
  GetHeroesPerRow,
  SetHeroesPerRow,
  GetSteamAccounts,
} from '../../wailsjs/go/main/App'
import { config, steam } from '../../wailsjs/go/models'
import { EventHeroesLayoutDataChanged, EventSteamAccountsChanged } from '../events'
import AccountCard from '../components/AccountCard'
import RelativeTime from '../components/RelativeTime'
import { AlertCircleIcon, GripIcon, MoreIcon, RefreshIcon, TrashIcon, XIcon } from '../components/Icons'

// Heroes per row constraints
const MIN_HEROES_PER_ROW = 1
const MAX_HEROES_PER_ROW = 50
const DEFAULT_HEROES_PER_ROW = 15

// Position name mapping
const positionNames: Record<string, string> = {
  '1': 'Carry',
  '2': 'Mid',
  '3': 'Offlane',
  '4': 'Support',
  '5': 'Hard Support',
}

interface DragState {
  draggedItem: config.PositionConfig
  originalPositions: config.PositionConfig[]
}

function HeroesLayoutPage() {
  // Loading state
  const [isLoading, setIsLoading] = useState(true)

  // Update state
  const [isUpdating, setIsUpdating] = useState(false)

  // Config files state
  const [files, setFiles] = useState<config.FileConfig[]>([])

  // Steam accounts state
  const [steamAccounts, setSteamAccounts] = useState<steam.SteamAccountView[]>([])

  // Positions state
  const [positions, setPositions] = useState<config.PositionConfig[]>([])

  // Drag and drop state
  const [dragState, setDragState] = useState<DragState | null>(null)

  // Error state
  const [error, setError] = useState<string | null>(null)

  // Heroes per row state
  const [heroesPerRow, setHeroesPerRowState] = useState<number>(DEFAULT_HEROES_PER_ROW)
  const [heroesPerRowInput, setHeroesPerRowInput] = useState<string>(DEFAULT_HEROES_PER_ROW.toString())

  // Menu state
  const [menuOpen, setMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    // Load initial states
    Promise.all([
      GetHeroesLayoutFiles().then(setFiles),
      GetPositions().then(setPositions),
      GetHeroesPerRow().then((value: number) => {
        setHeroesPerRowState(value)
        setHeroesPerRowInput(value.toString())
      }),
      GetSteamAccounts().then(setSteamAccounts),
    ]).catch(console.error).finally(() => setIsLoading(false))

    // Listen for background update notifications
    const offDataChanged = EventsOn(EventHeroesLayoutDataChanged, () => {
      GetHeroesLayoutFiles().then(setFiles).catch(console.error)
    })

    const offSteamChanged = EventsOn(EventSteamAccountsChanged, () => {
      GetSteamAccounts().then(setSteamAccounts).catch(console.error)
    })

    return () => {
      offDataChanged()
      offSteamChanged()
    }
  }, [])

  // Close menu on outside click
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false)
      }
    }

    if (menuOpen) {
      document.addEventListener('mousedown', handleClickOutside)
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [menuOpen])

  const handleUpdate = async () => {
    setIsUpdating(true)
    setError(null)
    try {
      await UpdateHeroesLayout()
      const [updatedFiles, updatedAccounts] = await Promise.all([
        GetHeroesLayoutFiles(),
        GetSteamAccounts(),
      ])
      setFiles(updatedFiles)
      setSteamAccounts(updatedAccounts)
    } catch (err) {
      console.error('Error updating heroes layout:', err)
      setError(`Failed to update heroes layout: ${err}`)
    } finally {
      setIsUpdating(false)
    }
  }

  const handleAddFile = async () => {
    setMenuOpen(false)
    setError(null)
    try {
      const path = await OpenFileDialog()
      if (path) {
        await AddHeroesLayoutFile(path)
        const updatedFiles = await GetHeroesLayoutFiles()
        setFiles(updatedFiles)
      }
    } catch (err) {
      console.error('Error adding file:', err)
      setError(`Failed to add file: ${err}`)
    }
  }

  const dismissError = () => {
    setError(null)
  }

  const handleHeroesPerRowChange = async (value: string) => {
    setHeroesPerRowInput(value)

    // Only update backend if value is valid
    const numValue = parseInt(value, 10)
    if (!isNaN(numValue) && numValue >= MIN_HEROES_PER_ROW && numValue <= MAX_HEROES_PER_ROW) {
      try {
        await SetHeroesPerRow(numValue)
        setHeroesPerRowState(numValue)
        setError(null)
      } catch (err) {
        console.error('Error setting heroes per row:', err)
        setError(`Failed to set heroes per row: ${err}`)
      }
    }
  }

  const handleHeroesPerRowBlur = () => {
    // On blur, reset to last valid value if input is invalid
    const numValue = parseInt(heroesPerRowInput, 10)
    if (isNaN(numValue) || numValue < MIN_HEROES_PER_ROW || numValue > MAX_HEROES_PER_ROW) {
      setHeroesPerRowInput(heroesPerRow.toString())
    }
  }

  const handleRemoveFile = async (filePath: string) => {
    try {
      await RemoveHeroesLayoutFile(filePath)
      const updatedFiles = await GetHeroesLayoutFiles()
      setFiles(updatedFiles)
    } catch (error) {
      console.error('Error removing file:', error)
    }
  }

  const handleToggleFileEnabled = async (filePath: string, enabled: boolean) => {
    try {
      await SetHeroesLayoutFileEnabled(filePath, enabled)
      const updatedFiles = await GetHeroesLayoutFiles()
      setFiles(updatedFiles)
    } catch (error) {
      console.error('Error toggling file:', error)
    }
  }

  const handleTogglePositionEnabled = async (id: string, enabled: boolean) => {
    try {
      await SetPositionEnabled(id, enabled)
      const updatedPositions = await GetPositions()
      setPositions(updatedPositions)
    } catch (error) {
      console.error('Error toggling position:', error)
    }
  }

  // Drag and drop handlers
  const handleDragStart = (e: React.DragEvent, position: config.PositionConfig) => {
    setDragState({
      draggedItem: position,
      originalPositions: [...positions],
    })
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', position.id)
  }

  const handleDragEnd = async () => {
    if (!dragState) return

    const { originalPositions } = dragState

    // Clear drag state first
    setDragState(null)

    // Check if order actually changed
    const orderChanged = positions.some((p, i) => p.id !== originalPositions[i].id)

    if (orderChanged) {
      try {
        await SetPositions(positions)
      } catch (error) {
        console.error('Error saving positions order:', error)
        // Revert on error
        setPositions(originalPositions)
      }
    }
  }

  const handleDragOver = (e: React.DragEvent, targetIndex: number) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'

    if (!dragState) return

    const currentIndex = positions.findIndex(p => p.id === dragState.draggedItem.id)
    if (currentIndex === -1 || currentIndex === targetIndex) return

    // Update positions directly during drag
    const newPositions = [...positions]
    newPositions.splice(currentIndex, 1)
    newPositions.splice(targetIndex, 0, dragState.draggedItem)
    setPositions(newPositions)
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
  }

  const getPositionName = (id: string) => {
    return positionNames[id] || `Position ${id}`
  }

  // Filter enabled accounts for display
  const enabledAccounts = steamAccounts.filter(a => a.enabled)

  const hasAccountsOrFiles = enabledAccounts.length > 0 || files.length > 0

  if (isLoading) {
    return (
      <div className="page">
        <div className="page-header">
          <div className="page-header-text">
            <h1 className="page-title">Heroes Layout</h1>
          </div>
        </div>
        <div className="loading-state">Loading...</div>
      </div>
    )
  }

  return (
    <div className="page">
      <div className="page-header">
        <div className="page-header-text">
          <h1 className="page-title">Heroes Layout</h1>
          <p className="page-description">Manage your heroes grid configuration and layout order</p>
        </div>
        <button
          className="btn btn-primary"
          onClick={handleUpdate}
          disabled={isUpdating}
        >
          <RefreshIcon />
          <span>{isUpdating ? 'Updating...' : 'Update Grids'}</span>
        </button>
      </div>

      {isUpdating && (
        <div className="progress-bar">
          <div className="progress-bar-inner"></div>
        </div>
      )}

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

        {/* Accounts Card (merged with Custom Files) */}
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">Accounts</h2>
            <div className="dropdown" ref={menuRef}>
              <button
                className="btn btn-icon"
                onClick={() => setMenuOpen(!menuOpen)}
                title="More options"
              >
                <MoreIcon />
              </button>
              {menuOpen && (
                <div className="dropdown-menu">
                  <button className="dropdown-item" onClick={handleAddFile}>
                    Add custom file...
                  </button>
                </div>
              )}
            </div>
          </div>
          <div className="card-body">
            {!hasAccountsOrFiles ? (
              <div className="empty-state">
                <p>No accounts enabled</p>
                <p className="empty-state-hint">Configure accounts in Steam settings</p>
              </div>
            ) : (
              <div className="file-list">
                {enabledAccounts.map((account) => (
                  <AccountCard key={account.steamId64} account={account} />
                ))}
                {files.map((file) => (
                  <div key={file.filePath} className={`file-card ${!file.enabled ? 'disabled' : ''} ${file.lastUpdateErrorMessage ? 'has-error' : ''}`}>
                    <div className="file-card-header">
                      <label className="toggle toggle-sm">
                        <input
                          type="checkbox"
                          checked={file.enabled}
                          onChange={(e) => handleToggleFileEnabled(file.filePath, e.target.checked)}
                        />
                        <span className="toggle-slider"></span>
                      </label>
                      <div className="file-card-title">
                        <span className="file-path" title={file.filePath}>{file.filePath}</span>
                      </div>
                      <button
                        className="btn btn-icon btn-danger"
                        onClick={() => handleRemoveFile(file.filePath)}
                        title="Remove file"
                      >
                        <TrashIcon />
                      </button>
                    </div>
                    <div className="file-card-footer">
                      <RelativeTime timestampMillis={file.lastUpdateTimestampMillis} prefix="Updated: " />
                      {file.lastUpdateErrorMessage && (
                        <span className="file-error">{file.lastUpdateErrorMessage}</span>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Positions Order Card */}
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">Positions Order</h2>
          </div>
          <div className="card-body">
            {positions.length === 0 ? (
              <div className="empty-state">
                <p>No positions configured</p>
              </div>
            ) : (
              <div className="list sortable-list">
                {positions.map((position, index) => (
                  <div
                    key={position.id}
                    className={`list-item draggable ${dragState?.draggedItem.id === position.id ? 'dragging' : ''} ${!position.enabled ? 'disabled' : ''}`}
                    draggable
                    onDragStart={(e) => handleDragStart(e, position)}
                    onDragEnd={handleDragEnd}
                    onDragOver={(e) => handleDragOver(e, index)}
                    onDrop={handleDrop}
                  >
                    <div className="list-item-content">
                      <span className="drag-handle">
                        <GripIcon />
                      </span>
                      <label className="toggle toggle-sm">
                        <input
                          type="checkbox"
                          checked={position.enabled}
                          onChange={(e) => {
                            e.stopPropagation()
                            handleTogglePositionEnabled(position.id, e.target.checked)
                          }}
                        />
                        <span className="toggle-slider"></span>
                      </label>
                      <span className="list-item-text">{getPositionName(position.id)}</span>
                    </div>
                  </div>
                ))}
              </div>
            )}
            <div className="card-hint">Drag to reorder, toggle to enable/disable positions</div>
          </div>
        </div>

        {/* Layout Settings Card */}
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">Layout Settings</h2>
          </div>
          <div className="card-body">
            <div className="setting-row">
              <div className="setting-info">
                <div className="setting-label">Heroes Per Row</div>
                <div className="setting-description">
                  Number of hero icons to display in each row ({MIN_HEROES_PER_ROW}-{MAX_HEROES_PER_ROW})
                </div>
              </div>
              <input
                type="number"
                className="select"
                min={MIN_HEROES_PER_ROW}
                max={MAX_HEROES_PER_ROW}
                required
                value={heroesPerRowInput}
                onChange={(e) => handleHeroesPerRowChange(e.target.value)}
                onBlur={handleHeroesPerRowBlur}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default HeroesLayoutPage
