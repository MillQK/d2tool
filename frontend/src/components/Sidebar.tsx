import { useState } from 'react'

export type PageId = 'heroesLayout' | 'startup' | 'updates'

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
}

// SVG Icons
const GridIcon = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect x="3" y="3" width="7" height="7" />
    <rect x="14" y="3" width="7" height="7" />
    <rect x="14" y="14" width="7" height="7" />
    <rect x="3" y="14" width="7" height="7" />
  </svg>
)

const RocketIcon = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 0 0-2.91-.09z" />
    <path d="M12 15l-3-3a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22.35 22.35 0 0 1-4 2z" />
    <path d="M9 12H4s.55-3.03 2-4c1.62-1.08 5 0 5 0" />
    <path d="M12 15v5s3.03-.55 4-2c1.08-1.62 0-5 0-5" />
  </svg>
)

const DownloadIcon = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
    <polyline points="7 10 12 15 17 10" />
    <line x1="12" y1="15" x2="12" y2="3" />
  </svg>
)

const ChevronIcon = ({ expanded }: { expanded: boolean }) => (
  <svg
    width="16"
    height="16"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    style={{ transform: expanded ? 'rotate(90deg)' : 'rotate(0deg)', transition: 'transform 0.2s ease' }}
  >
    <polyline points="9 18 15 12 9 6" />
  </svg>
)

function Sidebar({ activePage, onPageChange, updateAvailable }: SidebarProps) {
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
        { id: 'startup', label: 'Startup', icon: <RocketIcon /> },
        { id: 'updates', label: 'Updates', icon: <DownloadIcon /> },
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
