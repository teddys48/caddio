import type { User } from '../types';
import { api } from '../api/client';

let user = $state<User | undefined>(undefined);
let authenticated = $state<boolean>(false);
let setupRequired = $state<boolean>(false);
let loading = $state<boolean>(true);

export function useAuth() {
  return {
    get user() {
      return user;
    },
    get authenticated() {
      return authenticated;
    },
    get setupRequired() {
      return setupRequired;
    },
    get loading() {
      return loading;
    },
    async init() {
      loading = true;
      try {
        const state = await api.checkAuth();
        authenticated = state.authenticated;
        setupRequired = state.setup_required;
        user = state.user;
      } catch (err) {
        console.error('Failed checking authentication state:', err);
        authenticated = false;
        setupRequired = false;
        user = undefined;
      } finally {
        loading = false;
      }
    },
    setUser(u: User) {
      user = u;
      authenticated = true;
      setupRequired = false;
    },
    clear() {
      user = undefined;
      authenticated = false;
    },
  };
}
