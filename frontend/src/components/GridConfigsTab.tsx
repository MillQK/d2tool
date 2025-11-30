import { useEffect, useState } from 'react'
import {
  GetGridConfigPaths,
  AddGridConfigPath,
  RemoveGridConfigPath,
  OpenFileDialog,
} from '../wailsjs/go/main/App'

function GridConfigsTab() {
  const [paths, setPaths] = useState<string[]>([])

  useEffect(() => {
    loadPaths()
  }, [])

  const loadPaths = () => {
    GetGridConfigPaths().then(setPaths).catch(console.error)
  }

  const handleAddConfig = async () => {
    try {
      const path = await OpenFileDialog()
      if (path) {
        await AddGridConfigPath(path)
        loadPaths()
      }
    } catch (error) {
      console.error('Error adding config:', error)
    }
  }

  const handleRemove = async (index: number) => {
    try {
      await RemoveGridConfigPath(index)
      loadPaths()
    } catch (error) {
      console.error('Error removing config:', error)
    }
  }

  return (
    <div className="vbox">
      <h2 className="section-title">Heroes grid config paths</h2>

      <div className="list-container">
        {paths.length === 0 ? (
          <div className="empty-state">No config files added</div>
        ) : (
          paths.map((path, index) => (
            <div key={index} className="list-item">
              <span className="list-item-text" title={path}>
                {path}
              </span>
              <div className="list-item-actions">
                <button
                  className="button button-small button-danger"
                  onClick={() => handleRemove(index)}
                  title="Remove"
                >
                  Delete
                </button>
              </div>
            </div>
          ))
        )}
      </div>

      <button className="button" onClick={handleAddConfig}>
        Add Config
      </button>
    </div>
  )
}

export default GridConfigsTab
