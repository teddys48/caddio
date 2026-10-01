export type Theme = 'light' | 'dark' | 'system';

let currentTheme = $state<Theme>(
  (typeof localStorage !== 'undefined' && (localStorage.getItem('caddio_theme') as Theme)) || 'system'
);

function applyTheme(theme: Theme) {
  if (typeof document === 'undefined') return;
  const root = document.documentElement;

  const isDark =
    theme === 'dark' ||
    (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);

  if (isDark) {
    root.classList.add('dark');
  } else {
    root.classList.remove('dark');
  }
}

// Initial application
if (typeof window !== 'undefined') {
  applyTheme(currentTheme);

  // Listen for system theme changes if set to 'system'
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    if (currentTheme === 'system') {
      applyTheme('system');
    }
  });
}

export function useTheme() {
  return {
    get theme() {
      return currentTheme;
    },
    setTheme(newTheme: Theme) {
      currentTheme = newTheme;
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('caddio_theme', newTheme);
      }
      applyTheme(newTheme);
    },
  };
}
