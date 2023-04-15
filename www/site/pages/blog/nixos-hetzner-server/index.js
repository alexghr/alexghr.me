module.exports = {
  pageTitle: "NixOS VM on Hetzner with Flakes & Btrfs",
  url: "/blog/hetzner-nixos-server",
  extends: __dirname + "/../../../_templates/article",
  date: "2023-04-15",
  list: true,
  partials: {
    article: __dirname + "/post.md",
  },
  assets: [
    __dirname + "/iso.png",
    __dirname + "/snapshot.png",
    __dirname + "/add-server-snapshot.png",
  ],
};
