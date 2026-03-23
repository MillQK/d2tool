import { useEffect, useState } from 'react'
import { EventsOn } from '../wailsjs/runtime'
import { GetAppUpdateState, IsSteamPathValid } from '../wailsjs/go/main/App'
import Sidebar, { PageId } from './components/Sidebar'
import HeroesLayoutPage from './pages/HeroesLayoutPage'
import ProvidersPage from './pages/ProvidersPage'
import StartupPage from './pages/StartupPage'
import UpdatesPage from './pages/UpdatesPage'
import SteamPage from './pages/SteamPage'
import { main } from "../wailsjs/go/models"
import { EventAppUpdateDataChanged, EventSteamPathChanged } from './events'

function App() {
  const [activePage, setActivePage] = useState<PageId>('heroesLayout')
  const [appUpdateState, setAppUpdateState] = useState<main.AppUpdateState | null>(null)
  const [steamPathValid, setSteamPathValid] = useState(true)

  const refreshAppUpdateState = () => {
    return GetAppUpdateState().then(setAppUpdateState).catch(console.error)
  }

  const refreshSteamPathValid = () => {
    IsSteamPathValid().then(setSteamPathValid).catch(console.error)
  }

  useEffect(() => {
    refreshAppUpdateState()

    const offDataChanged = EventsOn(EventAppUpdateDataChanged, () => {
      refreshAppUpdateState()
    })

    return () => {
      offDataChanged()
    }
  }, [])

  useEffect(() => {
    refreshSteamPathValid()

    const offSteamPathChanged = EventsOn(EventSteamPathChanged, () => {
      refreshSteamPathValid()
    })

    return () => {
      offSteamPathChanged()
    }
  }, [])

  const renderPage = () => {
    switch (activePage) {
      case 'heroesLayout':
        return <HeroesLayoutPage />
      case 'steam':
        return <SteamPage />
      case 'providers':
        return <ProvidersPage />
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
      <Sidebar activePage={activePage} onPageChange={setActivePage} updateAvailable={appUpdateState?.updateAvailable} steamPathInvalid={!steamPathValid} />
      <main className="main-content">
        {renderPage()}
      </main>
    </div>
  )
}

export default App
