import { useEffect, useState } from 'react'
import { EventsOn } from '../wailsjs/runtime'
import {
  GetHomeState,
  UpdateHeroesGrid,
  GetGridConfigPaths,
  AddGridConfigPath,
  RemoveGridConfigPath,
  OpenFileDialog,
  GetPositionsOrder,
  SetPositionsOrder,
  HomeState,
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

const ClockIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="12" cy="12" r="10" />
    <polyline points="12 6 12 12 16 14" />
  </svg>
)

interface DragState {
  draggedItem: string
  originalPositions: string[]
}

function HeroesLayoutPage() {
  // Home state
  const [homeState, setHomeState] = useState<HomeState>({
    lastUpdateTime: 'Loading...',
    lastUpdateError: '',
    isUpdating: false,
  })

  // Grid configs state
  const [configPaths, setConfigPaths] = useState<string[]>([])

  // Positions state
  const [positions, setPositions] = useState<string[]>([])

  // Drag and drop state - simplified to just track dragged item and original order
  const [dragState, setDragState] = useState<DragState | null>(null)

  useEffect(() => {
    // Load initial states
    GetHomeState().then(setHomeState).catch(console.error)
    GetGridConfigPaths().then(setConfigPaths).catch(console.error)
    GetPositionsOrder().then(setPositions).catch(console.error)

    // Listen for update events
    const offStarted = EventsOn('heroesGridUpdateStarted', () => {
      setHomeState((prev) => ({ ...prev, isUpdating: true }))
    })

    const offFinished = EventsOn('heroesGridUpdateFinished', (newState: HomeState) => {
      setHomeState(newState)
    })

    return () => {
      offStarted()
      offFinished()
    }
  }, [])

  const handleUpdate = () => {
    UpdateHeroesGrid().catch(console.error)
  }

  const handleAddConfig = async () => {
    try {
      const path = await OpenFileDialog()
      if (path) {
        await AddGridConfigPath(path)
        const paths = await GetGridConfigPaths()
        setConfigPaths(paths)
      }
    } catch (error) {
      console.error('Error adding config:', error)
    }
  }

  const handleRemoveConfig = async (index: number) => {
    try {
      await RemoveGridConfigPath(index)
      const paths = await GetGridConfigPaths()
      setConfigPaths(paths)
    } catch (error) {
      console.error('Error removing config:', error)
    }
  }

  // Drag and drop handlers
  const handleDragStart = (e: React.DragEvent, position: string) => {
    setDragState({
      draggedItem: position,
      originalPositions: [...positions],
    })
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', position)
  }

  const handleDragEnd = async () => {
    if (!dragState) return

    const { originalPositions } = dragState

    // Clear drag state first
    setDragState(null)

    // Check if order actually changed
    const orderChanged = positions.some((p, i) => p !== originalPositions[i])

    if (orderChanged) {
      try {
        await SetPositionsOrder(positions)
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

    const currentIndex = positions.indexOf(dragState.draggedItem)
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

  const getPositionName = (position: string) => {
    return positionNames[position] || `Position ${position}`
  }

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
            <h2 className="card-title">Update Status</h2>
          </div>
          <div className="card-body">
            <div className="status-row">
              <div className="status-info">
                <div className="status-label">
                  <ClockIcon />
                  <span>Last Updated</span>
                </div>
                <div className="status-value">{homeState.lastUpdateTime}</div>
              </div>
              <button
                className="btn btn-primary"
                onClick={handleUpdate}
                disabled={homeState.isUpdating}
              >
                <RefreshIcon />
                <span>{homeState.isUpdating ? 'Updating...' : 'Update Now'}</span>
              </button>
            </div>

            {homeState.isUpdating && (
              <div className="progress-bar">
                <div className="progress-bar-inner"></div>
              </div>
            )}

            {homeState.lastUpdateError && (
              <div className="alert alert-error">
                <span>Error: {homeState.lastUpdateError}</span>
              </div>
            )}
          </div>
        </div>

        {/* Config Files Card */}
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">Config Files</h2>
            <button className="btn btn-secondary btn-sm" onClick={handleAddConfig}>
              <PlusIcon />
              <span>Add Config</span>
            </button>
          </div>
          <div className="card-body">
            {configPaths.length === 0 ? (
              <div className="empty-state">
                <p>No config files added yet</p>
                <p className="empty-state-hint">Click "Add Config" to select a heroes grid config file</p>
              </div>
            ) : (
              <div className="list">
                {configPaths.map((path, index) => (
                  <div key={index} className="list-item">
                    <div className="list-item-content">
                      <span className="list-item-text" title={path}>{path}</span>
                    </div>
                    <button
                      className="btn btn-icon btn-danger"
                      onClick={() => handleRemoveConfig(index)}
                      title="Remove config"
                    >
                      <TrashIcon />
                    </button>
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
            <p className="card-hint">Drag and drop to reorder positions</p>
            {positions.length === 0 ? (
              <div className="empty-state">
                <p>No positions configured</p>
              </div>
            ) : (
              <div className="list sortable-list">
                {positions.map((position, index) => (
                  <div
                    key={position}
                    className={`list-item draggable ${dragState?.draggedItem === position ? 'dragging' : ''}`}
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
                      <span className="list-item-text">{getPositionName(position)}</span>
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
