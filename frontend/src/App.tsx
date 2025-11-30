import { useState } from 'react'
import Sidebar, { PageId } from './components/Sidebar'
import HeroesLayoutPage from './pages/HeroesLayoutPage'
import StartupPage from './pages/StartupPage'
import UpdatesPage from './pages/UpdatesPage'

function App() {
  const [activePage, setActivePage] = useState<PageId>('heroesLayout')

  const renderPage = () => {
    switch (activePage) {
      case 'heroesLayout':
        return <HeroesLayoutPage />
      case 'startup':
        return <StartupPage />
      case 'updates':
        return <UpdatesPage />
      default:
        return <HeroesLayoutPage />
    }
  }

  return (
    <div className="app-layout">
      <Sidebar activePage={activePage} onPageChange={setActivePage} />
      <main className="main-content">
        {renderPage()}
      </main>
    </div>
  )
}

export default App
