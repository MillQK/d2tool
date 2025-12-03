import { useEffect, useState } from 'react'
import { EventsOn } from '../wailsjs/runtime'
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
  FileConfig,
  PositionConfig,
} from '../wailsjs/go/main/App'

// Position name mapping
const positionNames: Record<string, string> = {
  '1': 'Carry',
  '2': 'Mid',
  '3': 'Offlane',
  '4': 'Support',
  '5': 'Hard Support',
}

// Icons
const RefreshIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <polyline points="23 4 23 10 17 10" />
    <polyline points="1 20 1 14 7 14" />
    <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
  </svg>
)

const PlusIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <line x1="12" y1="5" x2="12" y2="19" />
    <line x1="5" y1="12" x2="19" y2="12" />
  </svg>
)

const TrashIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <polyline points="3 6 5 6 21 6" />
    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
  </svg>
)

const GripIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="9" cy="5" r="1" fill="currentColor" />
    <circle cx="9" cy="12" r="1" fill="currentColor" />
    <circle cx="9" cy="19" r="1" fill="currentColor" />
    <circle cx="15" cy="5" r="1" fill="currentColor" />
    <circle cx="15" cy="12" r="1" fill="currentColor" />
    <circle cx="15" cy="19" r="1" fill="currentColor" />
  </svg>
)

interface DragState {
  draggedItem: PositionConfig
  originalPositions: PositionConfig[]
}

function HeroesLayoutPage() {
  // Update state
  const [isUpdating, setIsUpdating] = useState(false)

  // Config files state
  const [files, setFiles] = useState<FileConfig[]>([])

  // Positions state
  const [positions, setPositions] = useState<PositionConfig[]>([])

  // Drag and drop state
  const [dragState, setDragState] = useState<DragState | null>(null)

  useEffect(() => {
    // Load initial states
    GetHeroesLayoutFiles().then(setFiles).catch(console.error)
    GetPositions().then(setPositions).catch(console.error)

    // Listen for background update notifications
    const offDataChanged = EventsOn('heroesLayoutDataChanged', () => {
      GetHeroesLayoutFiles().then(setFiles).catch(console.error)
    })

    return () => {
      offDataChanged()
    }
  }, [])

  const handleUpdate = async () => {
    setIsUpdating(true)
    try {
      const updatedFiles = await UpdateHeroesLayout()
      setFiles(updatedFiles)
    } catch (error) {
      console.error('Error updating heroes layout:', error)
    } finally {
      setIsUpdating(false)
    }
  }

  const handleAddFile = async () => {
    try {
      const path = await OpenFileDialog()
      if (path) {
        await AddHeroesLayoutFile(path)
        const updatedFiles = await GetHeroesLayoutFiles()
        setFiles(updatedFiles)
      }
    } catch (error) {
      console.error('Error adding file:', error)
    }
  }

  const handleRemoveFile = async (index: number) => {
    try {
      await RemoveHeroesLayoutFile(index)
      const updatedFiles = await GetHeroesLayoutFiles()
      setFiles(updatedFiles)
    } catch (error) {
      console.error('Error removing file:', error)
    }
  }

  const handleToggleFileEnabled = async (index: number, enabled: boolean) => {
    try {
      await SetHeroesLayoutFileEnabled(index, enabled)
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
  const handleDragStart = (e: React.DragEvent, position: PositionConfig) => {
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

  const formatLastUpdate = (timestampMillis: number) => {
    if (timestampMillis === 0) return 'Never'
    return new Date(timestampMillis).toLocaleString()
  }

  // Check if any file has errors
  const hasErrors = files.some(f => f.lastUpdateErrorMessage)

  return (
    <div className="page">
      <div className="page-header">
        <h1 className="page-title">Heroes Layout</h1>
        <p className="page-description">Manage your heroes grid configuration and layout order</p>
      </div>

      <div className="page-content">
        {/* Update Status Card */}
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">Update</h2>
            <button
              className="btn btn-primary btn-sm"
              onClick={handleUpdate}
              disabled={isUpdating}
            >
              <RefreshIcon />
              <span>{isUpdating ? 'Updating...' : 'Update Now'}</span>
            </button>
          </div>
          <div className="card-body">
            {isUpdating && (
              <div className="progress-bar">
                <div className="progress-bar-inner"></div>
              </div>
            )}

            {!isUpdating && hasErrors && (
              <div className="alert alert-error">
                <span>Some files failed to update. Check the file list below for details.</span>
              </div>
            )}

            {!isUpdating && !hasErrors && files.length > 0 && (
              <div className="alert alert-info">
                <span>All files are up to date.</span>
              </div>
            )}
          </div>
        </div>

        {/* Config Files Card */}
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">Config Files</h2>
            <button className="btn btn-secondary btn-sm" onClick={handleAddFile}>
              <PlusIcon />
              <span>Add File</span>
            </button>
          </div>
          <div className="card-body">
            {files.length === 0 ? (
              <div className="empty-state">
                <p>No config files added yet</p>
                <p className="empty-state-hint">Click "Add File" to select a heroes grid config file</p>
              </div>
            ) : (
              <div className="file-list">
                {files.map((file, index) => (
                  <div key={file.filePath} className={`file-card ${!file.enabled ? 'disabled' : ''} ${file.lastUpdateErrorMessage ? 'has-error' : ''}`}>
                    <div className="file-card-header">
                      <label className="toggle toggle-sm">
                        <input
                          type="checkbox"
                          checked={file.enabled}
                          onChange={(e) => handleToggleFileEnabled(index, e.target.checked)}
                        />
                        <span className="toggle-slider"></span>
                      </label>
                      <div className="file-card-title">
                        <span className="file-path" title={file.filePath}>{file.filePath}</span>
                        {file.attributes && Object.keys(file.attributes).length > 0 && (
                          <div className="file-attributes">
                            {Object.entries(file.attributes).map(([key, value]) => (
                              <span key={key} className="file-attribute">
                                <span className="file-attribute-key">{key}:</span> {value}
                              </span>
                            ))}
                          </div>
                        )}
                      </div>
                      <button
                        className="btn btn-icon btn-danger"
                        onClick={() => handleRemoveFile(index)}
                        title="Remove file"
                      >
                        <TrashIcon />
                      </button>
                    </div>
                    <div className="file-card-footer">
                      <span className="file-status">
                        {file.lastUpdateTimestampMillis > 0
                          ? `Updated: ${formatLastUpdate(file.lastUpdateTimestampMillis)}`
                          : 'Never updated'}
                      </span>
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
            <p className="card-hint">Drag to reorder, toggle to enable/disable positions</p>
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
          </div>
        </div>
      </div>
    </div>
  )
}

export default HeroesLayoutPage
