// Mock for login has been disabled — requests are now proxied to the real Go backend.
// The fake tokens and hardcoded GitHub avatar URLs were causing:
//   1. 401 errors on /api/admin (fake token rejected by real backend)
//   2. Wrong avatar display (GitHub URL instead of /avatar.png?username=xxx)
import { defineFakeRoute } from "vite-plugin-fake-server/client";
export default defineFakeRoute([]);
