// Mock for /refresh-token has been disabled — requests are now proxied to the real Go backend.
// The mock was returning fake tokens that the real backend couldn't validate.
import { defineFakeRoute } from "vite-plugin-fake-server/client";
export default defineFakeRoute([]);
