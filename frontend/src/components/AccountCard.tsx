import { steam } from '../../wailsjs/go/models'
import { formatLastUpdate } from '../utils/format'
import { UserIcon } from './Icons'

interface AccountCardProps {
  account: steam.SteamAccountView
  toggle?: {
    checked: boolean
    onChange: (enabled: boolean) => void
  }
}

function AccountCard({ account, toggle }: AccountCardProps) {
  return (
    <div className={`file-card ${toggle && !account.enabled ? 'disabled' : ''} ${account.lastUpdateErrorMessage ? 'has-error' : ''}`}>
      <div className="account-card-header">
        {toggle && (
          <label className="toggle toggle-sm">
            <input
              type="checkbox"
              checked={toggle.checked}
              onChange={(e) => toggle.onChange(e.target.checked)}
            />
            <span className="toggle-slider"></span>
          </label>
        )}
        <div className="account-avatar">
          {account.avatarBase64 ? (
            <img
              src={`data:image/png;base64,${account.avatarBase64}`}
              alt="avatar"
            />
          ) : (
            <UserIcon />
          )}
        </div>
        <div className="account-info">
          <span className="account-name">{account.personaName || account.accountName || account.steamId64}</span>
          {account.personaName && account.accountName && (
            <span className="account-username">{account.accountName}</span>
          )}
        </div>
      </div>
      <div className="file-card-footer">
        <span className="file-status">
          {account.lastUpdateTimestampMillis > 0
            ? `Updated: ${formatLastUpdate(account.lastUpdateTimestampMillis)}`
            : 'Never updated'}
        </span>
        {account.lastUpdateErrorMessage && (
          <span className="file-error">{account.lastUpdateErrorMessage}</span>
        )}
      </div>
    </div>
  )
}

export default AccountCard
