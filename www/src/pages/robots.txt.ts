import type { APIRoute } from "astro";

export const get: APIRoute = async (context) => {
  return new Response(`\
User-agent: *
Disallow: /_astro/
Sitemap: ${new URL("/sitemap-index.xml", context.site).href}
`);
};
