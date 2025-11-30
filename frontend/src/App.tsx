import { useState } from 'react'
import HomeTab from './components/HomeTab'
import GridConfigsTab from './components/GridConfigsTab'
import PositionsOrderTab from './components/PositionsOrderTab'
import StartupTab from './components/StartupTab'
import AppUpdateTab from './components/AppUpdateTab'

type TabId = 'home' | 'gridConfigs' | 'positionsOrder' | 'startup' | 'appUpdate'

interface Tab {
  id: TabId
  label: string
  component: React.ReactNode
}

function App() {
  const [activeTab, setActiveTab] = useState<TabId>('home')

  const tabs: Tab[] = [
    { id: 'home', label: 'Home', component: <HomeTab /> },
    { id: 'gridConfigs', label: 'Grid configs', component: <GridConfigsTab /> },
    { id: 'positionsOrder', label: 'Positions order', component: <PositionsOrderTab /> },
    { id: 'startup', label: 'Startup', component: <StartupTab /> },
    { id: 'appUpdate', label: 'App update', component: <AppUpdateTab /> },
  ]

  return (
    <div className="app-container">
      <div className="tabs-header">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            className={`tab-button ${activeTab === tab.id ? 'active' : ''}`}
            onClick={() => setActiveTab(tab.id)}
          >
            {tab.label}
          </button>
        ))}
      </div>
      <div className="tab-content">
        {tabs.find((tab) => tab.id === activeTab)?.component}
      </div>
    </div>
  )
}

export default App
