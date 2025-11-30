import { useEffect, useState } from 'react'
import { EventsOn } from '../wailsjs/runtime'
import { GetHomeState, UpdateHeroesGrid, HomeState } from '../wailsjs/go/main/App'

function HomeTab() {
  const [state, setState] = useState<HomeState>({
    lastUpdateTime: 'Loading...',
    lastUpdateError: '',
    isUpdating: false,
  })

  useEffect(() => {
    // Load initial state
    GetHomeState().then(setState).catch(console.error)

    // Listen for update events
    const offStarted = EventsOn('heroesGridUpdateStarted', () => {
      setState((prev) => ({ ...prev, isUpdating: true }))
    })

    const offFinished = EventsOn('heroesGridUpdateFinished', (newState: HomeState) => {
      setState(newState)
    })

    return () => {
      offStarted()
      offFinished()
    }
  }, [])

  const handleUpdate = () => {
    UpdateHeroesGrid().catch(console.error)
  }

  return (
    <div className="vbox">
      <p className="label label-bold">Last update time: {state.lastUpdateTime}</p>

      {state.lastUpdateError && (
        <p className="error-text">Error: {state.lastUpdateError}</p>
      )}

      <button
        className="button"
        onClick={handleUpdate}
        disabled={state.isUpdating}
      >
        Update heroes grid configs
      </button>

      {state.isUpdating && (
        <div className="progress-bar">
          <div className="progress-bar-inner"></div>
        </div>
      )}
    </div>
  )
}

export default HomeTab
