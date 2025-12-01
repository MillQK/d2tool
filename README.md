# D2Tool - Dota 2 Hero Grid Tool

D2Tool is a desktop application that helps you manage your Dota 2 hero grid configurations. It automatically updates your Dota 2 heroes grid config files periodically in the background, organizing heroes based on their performance data from Dota 2 Pro Tracker.

## Features

- **Automatic Grid Generation**: Creates hero grid layouts based on hero ratings and match counts from Dota 2 Pro Tracker
- **Position-Based Organization**: Organizes heroes by their positions (Carry, Mid, Offlane, Support, Hard Support) with customizable order
- **Multiple Account Support**: Automatically finds and manages grid configs for all Steam accounts on your computer
- **Per-File Management**: Enable/disable individual config files, view last update time and errors for each file
- **Position Toggle**: Enable or disable specific positions to customize which roles appear in your grid
- **Drag & Drop Reordering**: Easily reorder positions by dragging them in the interface
- **Background Operation**: Runs in the background and updates grids periodically (every hour)
- **Auto-Updates**: Automatically checks for application updates on startup
- **Startup Integration**: Option to run automatically when your computer starts (Windows and macOS supported)
- **Modern Interface**: Clean, dark-themed UI with sidebar navigation

## How It Works

D2Tool fetches hero statistics from Dota 2 Pro Tracker, then:
1. Finds your Steam installation and locates all hero grid config files
2. Removes any previously generated D2Tool configurations
3. Creates new hero grid layouts organized by enabled positions and performance metrics
4. Saves the updated configurations back to your Dota 2 config files

## Installation

### Prerequisites

- Go 1.23 or later
- Node.js 18 or later
- Wails CLI v2
- Steam and Dota 2 installed

### Installing Wails

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

For platform-specific dependencies, see the [Wails Installation Guide](https://wails.io/docs/gettingstarted/installation).

### Building the Application

1. Clone the repository
2. Install frontend dependencies:
   ```bash
   cd frontend
   npm install
   cd ..
   ```
3. Build the application:
   ```bash
   wails build
   ```

The built binary will be in the `build/bin` directory.

### Development Mode

To run in development mode with hot reload:

```bash
wails dev
```

### Pre-built Binaries

Pre-built binaries for Windows and macOS are available in the [releases](https://github.com/MillQK/d2tool/releases) section.

## Usage

### Running the Application

Run the built binary or use the pre-built release for your platform.

Command line options:
- `--minimized` - Start the application minimized (useful for startup)

### Heroes Layout Page

The main page for managing your hero grid configurations:

- **Update Section**: Manually trigger an update of all enabled config files
- **Config Files**:
  - View all discovered hero grid config files with their attributes (account name, Steam ID)
  - Enable/disable individual files
  - See last update time and any errors for each file
  - Add custom config files or remove existing ones
- **Positions Order**:
  - Drag and drop to reorder positions
  - Toggle positions on/off to control which roles appear in your grid

### Startup Page

Configure whether D2Tool runs automatically when your computer starts.

### Updates Page

- Check for application updates
- Download and install new versions
- Enable/disable automatic update checks

## Troubleshooting

### Steam Path Not Found

If D2Tool cannot find your Steam installation:
1. Click "Add File" and manually navigate to your hero grid config file
2. The file is typically located at:
   - Windows: `C:\Program Files (x86)\Steam\userdata\<your-steam-id>\570\remote\cfg\hero_grid_config.json`
   - macOS: `~/Library/Application Support/Steam/userdata/<your-steam-id>/570/remote/cfg/hero_grid_config.json`

### Changes Not Appearing in Dota 2

If your updated grid layouts don't appear in Dota 2:
1. Make sure Dota 2 is closed when running D2Tool
2. Restart Dota 2 after running D2Tool
3. In Dota 2, check the "Heroes" tab and layouts there to see your updated layouts

### Logs

D2Tool creates a `d2tool.log` file in the same directory as the executable for debugging purposes.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
