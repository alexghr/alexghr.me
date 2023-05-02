import { getCollection } from "astro:content";

export const sortedBlogPosts = async () =>
  (await getCollection("blog")).sort(
    (a, b) => b.data.pubDate.valueOf() - a.data.pubDate.valueOf()
  );

export const recentBlogPosts = async () =>
  (await sortedBlogPosts()).slice(0, 5);
