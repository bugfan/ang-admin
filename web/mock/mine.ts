// Mock for /mine has been disabled — requests are now proxied to the real Go backend.
// The mock was returning a hardcoded GitHub avatar URL instead of the real user data.
import { defineFakeRoute } from "vite-plugin-fake-server/client";
export default defineFakeRoute([]);
