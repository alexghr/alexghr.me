const base = "/asset/1fd137b1fff7463f8019d8114a8d2f34b2b5859c0746b0a65fd14c9a48c5a7d0/es";

const { default: hljs } = await import(`${base}/highlight.min.js`);
await Promise.all(
  ["bash", "shell", "javascript", "nix", "css", "yaml"].map(async name => {
    const { default: grammar } =
      await import(`${base}/languages/${name}.min.js`);

    hljs.registerLanguage(name, grammar);
  })
);

hljs.highlightAll();
