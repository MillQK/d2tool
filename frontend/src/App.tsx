import { useEffect, useState } from 'react'
import { EventsOn } from './wailsjs/runtime'
import { GetAppUpdateState, AppUpdateState } from './wailsjs/go/main/App'
import Sidebar, { PageId } from './components/Sidebar'
import HeroesLayoutPage from './pages/HeroesLayoutPage'
import StartupPage from './pages/StartupPage'
import UpdatesPage from './pages/UpdatesPage'

function App() {
  const [activePage, setActivePage] = useState<PageId>('heroesLayout')
  const [appUpdateState, setAppUpdateState] = useState<AppUpdateState | null>(null)

  const refreshAppUpdateState = () => {
    return GetAppUpdateState().then(setAppUpdateState).catch(console.error)
  }

  useEffect(() => {
    refreshAppUpdateState()

    const offDataChanged = EventsOn('appUpdateDataChanged', () => {
      refreshAppUpdateState()
    })

    return () => {
      offDataChanged()
    }
  }, [])

  const renderPage = () => {
    switch (activePage) {
      case 'heroesLayout':
        return <HeroesLayoutPage />
      case 'startup':
        return <StartupPage />
      case 'updates':
        return <UpdatesPage state={appUpdateState} onStateChange={refreshAppUpdateState} />
      default:
        return <HeroesLayoutPage />
    }
  }

  return (
    <div className="app-layout">
      <Sidebar activePage={activePage} onPageChange={setActivePage} updateAvailable={appUpdateState?.updateAvailable} />
      <main className="main-content">
        {renderPage()}
      </main>
    </div>
  )
}

export default App
