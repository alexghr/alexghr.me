import { defineConfig } from "astro/config";
import mdx from "@astrojs/mdx";

import sitemap from "@astrojs/sitemap";

if (!process.env.SITE) {
  throw new Error("Missing SITE env var");
}

// https://astro.build/config
export default defineConfig({
  site: process.env.SITE,
  integrations: [mdx(), sitemap()],
  experimental: {
    assets: true,
  },
});
