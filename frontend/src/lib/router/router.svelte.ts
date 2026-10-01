export type Route =
  | '/setup'
  | '/login'
  | '/dashboard'
  | '/audit'
  | '/sites'
  | '/config'
  | '/config/history'
  | '/logs'
  | '/history'
  | '/settings';

let currentRoute = $state<Route>((window.location.pathname as Route) || '/dashboard');

if (typeof window !== 'undefined') {
  window.addEventListener('popstate', () => {
    currentRoute = (window.location.pathname as Route) || '/dashboard';
  });
}

export function useRouter() {
  return {
    get route(): Route {
      return currentRoute;
    },
    navigate(to: Route) {
      if (currentRoute !== to) {
        currentRoute = to;
        window.history.pushState({}, '', to);
      }
    },
  };
}
