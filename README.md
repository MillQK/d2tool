# D2Tool - Dota 2 Hero Grid Tool

D2Tool is a GUI application that helps you manage your Dota 2 hero grid configurations. It automatically updates your Dota 2 heroes grid config files periodically in the background, organizing heroes based on their performance data from Dota 2 Pro Tracker.

![D2Tool Screenshot](d2tool-icon.png)

## Features

- **Automatic Grid Generation**: Creates hero grid layouts based on hero ratings and match counts from Dota 2 Pro Tracker
- **Position-Based Organization**: Organizes heroes by their positions (1-5) and performance metrics
- **Multiple Account Support**: Finds and updates grid configs for all Steam accounts on your computer
- **Background Operation**: Can run in the background and update grids periodically
- **Startup Integration**: Option to run automatically when your computer starts (Windows supported)
- **User-Friendly Interface**: Simple GUI with main and settings panels

## How It Works

D2Tool fetches hero statistics from Dota 2 Pro Tracker, then:
1. Finds your Steam installation and locates all hero grid config files
2. Removes any previously generated D2Tool configurations
3. Creates new hero grid layouts organized by position and performance metrics
4. Saves the updated configurations back to your Dota 2 config files

## Installation

### Prerequisites

- Go 1.23 or later
- Fyne library
- Fyne-cross library
- Steam and Dota 2 installed

### Installing

To install the Fyne library, run:

```bash
go get fyne.io/fyne/v2 
go get github.com/fyne-io/fyne-cross 
go mod tidy
```

You may need additional dependencies for Fyne. See the [Fyne Getting Started guide](https://developer.fyne.io/started/) for platform-specific requirements.

### Building the Application

To build the application, run:

```bash
go build
```

### Pre-built Binaries

Pre-built binaries for Windows are available in the releases section.

## Usage

### Running the Application

To run the application, simply execute the built binary:

```bash
./d2tool
```

On Windows, you can double-click the executable file.

### Main Panel

The main panel allows you to:

1. View and manage hero grid config files
2. Add new config files manually
3. Find config files automatically using Steam path
4. Select which positions (1-5) to include in the layouts
5. Run the hero grid generation process

### Settings Panel

The settings panel allows you to:

1. Enable or disable run on startup (Windows only)

## Troubleshooting

### Steam Path Not Found

If D2Tool cannot find your Steam installation:
1. Click "Add Config" and manually navigate to your hero grid config file
2. The file is typically located at: `C:\Program Files (x86)\Steam\userdata\<your-steam-id>\570\remote\cfg\hero_grid_config.json`

### Changes Not Appearing in Dota 2

If your updated grid layouts don't appear in Dota 2:
1. Make sure Dota 2 is closed when running D2Tool
2. Restart Dota 2 after running D2Tool
3. In Dota 2, check the "Grid Editor" to see your updated layouts

## License

This project is licensed under the MIT License - see the LICENSE file for details.
