import { useEffect, useState } from 'react'
import {
  GetPositionsOrder,
  MovePositionUp,
  MovePositionDown,
} from '../wailsjs/go/main/App'

function PositionsOrderTab() {
  const [positions, setPositions] = useState<string[]>([])

  useEffect(() => {
    GetPositionsOrder().then(setPositions).catch(console.error)
  }, [])

  const handleMoveUp = async (index: number) => {
    try {
      const newPositions = await MovePositionUp(index)
      setPositions(newPositions)
    } catch (error) {
      console.error('Error moving position up:', error)
    }
  }

  const handleMoveDown = async (index: number) => {
    try {
      const newPositions = await MovePositionDown(index)
      setPositions(newPositions)
    } catch (error) {
      console.error('Error moving position down:', error)
    }
  }

  return (
    <div className="vbox">
      <h2 className="section-title">Heroes grid positions order</h2>

      <div className="list-container">
        {positions.map((position, index) => (
          <div key={index} className="list-item">
            <span className="list-item-text">Position {position}</span>
            <div className="list-item-actions">
              <div className="move-buttons">
                <button
                  className="button button-small move-button"
                  onClick={() => handleMoveUp(index)}
                  disabled={index === 0}
                  title="Move up"
                >
                  ▲
                </button>
                <button
                  className="button button-small move-button"
                  onClick={() => handleMoveDown(index)}
                  disabled={index === positions.length - 1}
                  title="Move down"
                >
                  ▼
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default PositionsOrderTab
