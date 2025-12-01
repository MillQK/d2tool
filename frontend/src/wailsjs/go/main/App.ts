// Wails bindings for Go App methods

export interface FileConfig {
  filePath: string;
  enabled: boolean;
  attributes: Record<string, string>;
  lastUpdateTimestampMillis: number;
  lastUpdateErrorMessage: string;
}

export interface PositionConfig {
  id: string;
  enabled: boolean;
}

export interface AppUpdateState {
  currentVersion: string;
  latestVersion: string;
  lastCheckTime: string;
  updateAvailable: boolean;
  isCheckingForUpdate: boolean;
  isDownloadingUpdate: boolean;
  autoUpdateEnabled: boolean;
}

// Declare the window.go object
declare global {
  interface Window {
    go: {
      main: {
        App: {
          // Heroes Layout Update
          GetIsUpdatingLayout(): Promise<boolean>;
          UpdateHeroesLayout(): Promise<void>;

          // Heroes Layout Files
          GetHeroesLayoutFiles(): Promise<FileConfig[]>;
          AddHeroesLayoutFile(path: string): Promise<void>;
          RemoveHeroesLayoutFile(index: number): Promise<void>;
          SetHeroesLayoutFileEnabled(index: number, enabled: boolean): Promise<void>;
          OpenFileDialog(): Promise<string>;

          // Positions
          GetPositions(): Promise<PositionConfig[]>;
          SetPositions(positions: PositionConfig[]): Promise<void>;
          SetPositionEnabled(id: string, enabled: boolean): Promise<void>;

          // Startup
          GetStartupEnabled(): Promise<boolean>;
          SetStartupEnabled(enabled: boolean): Promise<void>;
          IsStartupSupported(): Promise<boolean>;

          // App Update
          GetAppUpdateState(): Promise<AppUpdateState>;
          CheckForAppUpdate(): Promise<void>;
          DownloadAppUpdate(): Promise<void>;
          GetAutoUpdateEnabled(): Promise<boolean>;
          SetAutoUpdateEnabled(enabled: boolean): Promise<void>;
        };
      };
    };
  }
}

// --- Heroes Layout Update ---

export function GetIsUpdatingLayout(): Promise<boolean> {
  return window.go.main.App.GetIsUpdatingLayout();
}

export function UpdateHeroesLayout(): Promise<void> {
  return window.go.main.App.UpdateHeroesLayout();
}

// --- Heroes Layout Files ---

export function GetHeroesLayoutFiles(): Promise<FileConfig[]> {
  return window.go.main.App.GetHeroesLayoutFiles();
}

export function AddHeroesLayoutFile(path: string): Promise<void> {
  return window.go.main.App.AddHeroesLayoutFile(path);
}

export function RemoveHeroesLayoutFile(index: number): Promise<void> {
  return window.go.main.App.RemoveHeroesLayoutFile(index);
}

export function SetHeroesLayoutFileEnabled(index: number, enabled: boolean): Promise<void> {
  return window.go.main.App.SetHeroesLayoutFileEnabled(index, enabled);
}

export function OpenFileDialog(): Promise<string> {
  return window.go.main.App.OpenFileDialog();
}

// --- Positions ---

export function GetPositions(): Promise<PositionConfig[]> {
  return window.go.main.App.GetPositions();
}

export function SetPositions(positions: PositionConfig[]): Promise<void> {
  return window.go.main.App.SetPositions(positions);
}

export function SetPositionEnabled(id: string, enabled: boolean): Promise<void> {
  return window.go.main.App.SetPositionEnabled(id, enabled);
}

// --- Startup ---

export function GetStartupEnabled(): Promise<boolean> {
  return window.go.main.App.GetStartupEnabled();
}

export function SetStartupEnabled(enabled: boolean): Promise<void> {
  return window.go.main.App.SetStartupEnabled(enabled);
}

export function IsStartupSupported(): Promise<boolean> {
  return window.go.main.App.IsStartupSupported();
}

// --- App Update ---

export function GetAppUpdateState(): Promise<AppUpdateState> {
  return window.go.main.App.GetAppUpdateState();
}

export function CheckForAppUpdate(): Promise<void> {
  return window.go.main.App.CheckForAppUpdate();
}

export function DownloadAppUpdate(): Promise<void> {
  return window.go.main.App.DownloadAppUpdate();
}

export function GetAutoUpdateEnabled(): Promise<boolean> {
  return window.go.main.App.GetAutoUpdateEnabled();
}

export function SetAutoUpdateEnabled(enabled: boolean): Promise<void> {
  return window.go.main.App.SetAutoUpdateEnabled(enabled);
}
