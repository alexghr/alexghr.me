import rss from "@astrojs/rss";
import type { APIRoute } from "astro";
import { getCollection } from "astro:content";
import { SITE_DESCRIPTION, SITE_TITLE } from "../consts";

export const get: APIRoute = async (context) => {
  const posts = await getCollection("blog");
  const { body } = await rss({
    title: SITE_TITLE,
    description: SITE_DESCRIPTION,
    site: context.site?.href ?? "",
    customData: `<language>en-gb</language>`,
    items: posts.map((post) => ({
      ...post.data,
      link: `/blog/${post.slug}/`,
    })),
    trailingSlash: true,
  });

  return new Response(body, {
    headers: {
      "content-type": "application/rss+xml",
    },
  });
};
