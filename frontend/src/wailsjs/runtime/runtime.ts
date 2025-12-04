// Runtime stubs - these will be provided by Wails at runtime
declare global {
  interface Window {
    runtime: {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      EventsOn: (eventName: string, callback: (...data: any[]) => void) => () => void;
      EventsOff: (eventName: string, ...additionalEventNames: string[]) => void;
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      EventsOnce: (eventName: string, callback: (...data: any[]) => void) => () => void;
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      EventsOnMultiple: (eventName: string, callback: (...data: any[]) => void, maxCallbacks: number) => () => void;
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      EventsEmit: (eventName: string, ...data: any[]) => void;
      WindowReload: () => void;
      WindowReloadApp: () => void;
      WindowSetAlwaysOnTop: (b: boolean) => void;
      WindowSetSystemDefaultTheme: () => void;
      WindowSetLightTheme: () => void;
      WindowSetDarkTheme: () => void;
      WindowCenter: () => void;
      WindowSetTitle: (title: string) => void;
      WindowFullscreen: () => void;
      WindowUnfullscreen: () => void;
      WindowIsFullscreen: () => Promise<boolean>;
      WindowSetSize: (width: number, height: number) => void;
      WindowGetSize: () => Promise<{ w: number; h: number }>;
      WindowSetMaxSize: (width: number, height: number) => void;
      WindowSetMinSize: (width: number, height: number) => void;
      WindowSetPosition: (x: number, y: number) => void;
      WindowGetPosition: () => Promise<{ x: number; y: number }>;
      WindowHide: () => void;
      WindowShow: () => void;
      WindowMaximise: () => void;
      WindowToggleMaximise: () => void;
      WindowUnmaximise: () => void;
      WindowIsMaximised: () => Promise<boolean>;
      WindowMinimise: () => void;
      WindowUnminimise: () => void;
      WindowIsMinimised: () => Promise<boolean>;
      WindowIsNormal: () => Promise<boolean>;
      WindowSetBackgroundColour: (r: number, g: number, b: number, a: number) => void;
      BrowserOpenURL: (url: string) => void;
      Quit: () => void;
      Environment: () => Promise<{
        buildType: string;
        platform: string;
        arch: string;
      }>;
    };
  }
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function EventsOn(eventName: string, callback: (...data: any[]) => void): () => void {
  return window.runtime.EventsOn(eventName, callback);
}

export function EventsOff(eventName: string, ...additionalEventNames: string[]): void {
  window.runtime.EventsOff(eventName, ...additionalEventNames);
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function EventsOnce(eventName: string, callback: (...data: any[]) => void): () => void {
  return window.runtime.EventsOnce(eventName, callback);
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function EventsOnMultiple(eventName: string, callback: (...data: any[]) => void, maxCallbacks: number): () => void {
  return window.runtime.EventsOnMultiple(eventName, callback, maxCallbacks);
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function EventsEmit(eventName: string, ...data: any[]): void {
  window.runtime.EventsEmit(eventName, ...data);
}

export function WindowReload(): void {
  window.runtime.WindowReload();
}

export function WindowReloadApp(): void {
  window.runtime.WindowReloadApp();
}

export function WindowSetAlwaysOnTop(b: boolean): void {
  window.runtime.WindowSetAlwaysOnTop(b);
}

export function WindowSetSystemDefaultTheme(): void {
  window.runtime.WindowSetSystemDefaultTheme();
}

export function WindowSetLightTheme(): void {
  window.runtime.WindowSetLightTheme();
}

export function WindowSetDarkTheme(): void {
  window.runtime.WindowSetDarkTheme();
}

export function WindowCenter(): void {
  window.runtime.WindowCenter();
}

export function WindowSetTitle(title: string): void {
  window.runtime.WindowSetTitle(title);
}

export function WindowFullscreen(): void {
  window.runtime.WindowFullscreen();
}

export function WindowUnfullscreen(): void {
  window.runtime.WindowUnfullscreen();
}

export function WindowIsFullscreen(): Promise<boolean> {
  return window.runtime.WindowIsFullscreen();
}

export function WindowSetSize(width: number, height: number): void {
  window.runtime.WindowSetSize(width, height);
}

export function WindowGetSize(): Promise<{ w: number; h: number }> {
  return window.runtime.WindowGetSize();
}

export function WindowSetMaxSize(width: number, height: number): void {
  window.runtime.WindowSetMaxSize(width, height);
}

export function WindowSetMinSize(width: number, height: number): void {
  window.runtime.WindowSetMinSize(width, height);
}

export function WindowSetPosition(x: number, y: number): void {
  window.runtime.WindowSetPosition(x, y);
}

export function WindowGetPosition(): Promise<{ x: number; y: number }> {
  return window.runtime.WindowGetPosition();
}

export function WindowHide(): void {
  window.runtime.WindowHide();
}

export function WindowShow(): void {
  window.runtime.WindowShow();
}

export function WindowMaximise(): void {
  window.runtime.WindowMaximise();
}

export function WindowToggleMaximise(): void {
  window.runtime.WindowToggleMaximise();
}

export function WindowUnmaximise(): void {
  window.runtime.WindowUnmaximise();
}

export function WindowIsMaximised(): Promise<boolean> {
  return window.runtime.WindowIsMaximised();
}

export function WindowMinimise(): void {
  window.runtime.WindowMinimise();
}

export function WindowUnminimise(): void {
  window.runtime.WindowUnminimise();
}

export function WindowIsMinimised(): Promise<boolean> {
  return window.runtime.WindowIsMinimised();
}

export function WindowIsNormal(): Promise<boolean> {
  return window.runtime.WindowIsNormal();
}

export function WindowSetBackgroundColour(r: number, g: number, b: number, a: number): void {
  window.runtime.WindowSetBackgroundColour(r, g, b, a);
}

export function BrowserOpenURL(url: string): void {
  window.runtime.BrowserOpenURL(url);
}

export function Quit(): void {
  window.runtime.Quit();
}

export function Environment(): Promise<{
  buildType: string;
  platform: string;
  arch: string;
}> {
  return window.runtime.Environment();
}
