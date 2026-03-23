import { useState } from 'react'
import { ChevronIcon, DownloadIcon, GridIcon, PlugIcon, RocketIcon, SteamIcon } from './Icons'

export type PageId = 'heroesLayout' | 'providers' | 'steam' | 'startup' | 'updates'

interface SidebarSection {
  id: string
  title: string
  items: SidebarItem[]
}

interface SidebarItem {
  id: PageId
  label: string
  icon: React.ReactNode
}

interface SidebarProps {
  activePage: PageId
  onPageChange: (page: PageId) => void
  updateAvailable?: boolean
  steamPathInvalid?: boolean
}

function Sidebar({ activePage, onPageChange, updateAvailable, steamPathInvalid }: SidebarProps) {
  const [expandedSections, setExpandedSections] = useState<Record<string, boolean>>({
    features: true,
    settings: true,
  })

  const sections: SidebarSection[] = [
    {
      id: 'features',
      title: 'Features',
      items: [
        { id: 'heroesLayout', label: 'Heroes Layout', icon: <GridIcon /> },
      ],
    },
    {
      id: 'settings',
      title: 'Settings',
      items: [
        { id: 'steam', label: 'Steam', icon: <SteamIcon /> },
        { id: 'providers', label: 'Providers', icon: <PlugIcon /> },
        { id: 'startup', label: 'Startup', icon: <RocketIcon /> },
        { id: 'updates', label: 'Updates', icon: <DownloadIcon size={20} /> },
      ],
    },
  ]

  const toggleSection = (sectionId: string) => {
    setExpandedSections((prev) => ({
      ...prev,
      [sectionId]: !prev[sectionId],
    }))
  }

  return (
    <nav className="sidebar">
      <div className="sidebar-header">
        <h1 className="sidebar-title">D2Tool</h1>
      </div>

      <div className="sidebar-content">
        {sections.map((section) => (
          <div key={section.id} className="sidebar-section">
            <button
              className="sidebar-section-header"
              onClick={() => toggleSection(section.id)}
            >
              <ChevronIcon expanded={expandedSections[section.id]} />
              <span className="sidebar-section-title">{section.title}</span>
            </button>

            {expandedSections[section.id] && (
              <div className="sidebar-section-items">
                {section.items.map((item) => (
                  <button
                    key={item.id}
                    className={`sidebar-item ${activePage === item.id ? 'active' : ''}`}
                    onClick={() => onPageChange(item.id)}
                  >
                    <span className="sidebar-item-icon">{item.icon}</span>
                    <span className="sidebar-item-label">{item.label}</span>
                    {item.id === 'updates' && updateAvailable && (
                      <span className="sidebar-item-badge" />
                    )}
                    {item.id === 'steam' && steamPathInvalid && (
                      <span className="sidebar-item-badge sidebar-item-badge-warning" />
                    )}
                  </button>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>
    </nav>
  )
}

export default Sidebar
