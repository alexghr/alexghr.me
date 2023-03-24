[<< Back to list](/blog)

# {{ pageTitle }}

I’ve been thinking about resurrecting this website. Its tech stack is pretty ancient by this point: it’s a static site built form a bunch of Markdown files and `.js` spec files (instead of frontmatter 🤦) with a custom build process that generates HTML files (read more about it [here](/blog/rubber-ducking-a-blog) and [here](/blog/rubber-ducking-a-blog-2)). There's no support for bundling Javascript, no image optimizations and no caching. It's pretty bare bones.

Almost a year ago I moved all of my personal devices to Nix (and NixOS) and I liked it so much that I migrated the one VM I had hosting this site to NixOS as well 😄. Unfortunately my _inexperience_ with Nix showed and rather than using the platform, I enabled [Kubernetes](https://github.com/alexghr/nix/blob/538f9831a3dd051de49e0d4e64faba58fbdbe7ee/hosts/nix-1/configuration.nix#L103-L104) to manage the services. And that was fine, mostly, but it created a bunch of unnecessary work for me: on top of maintaining a NixOS VM now I also had to effectively admin a Kubernetes cluster. I can use Kubernetes as an app developer, to deploy services and pods and manage databases, but I don't know enough of how it works to confidently say my cluster is secure.

I want to simplify all that and not only refresh the site itself (moving away from custom build scripts and templates) but also reduce the maintenance burden. Some requirements for the new site:

1. Easy deployment solution. Ideally everything would be managed via Nix (maybe the site could be a Nix flake and deploying it would just mean updating the commit the flake points to?)
2. Modern stack. Ideally it wouldn't be based on React/Next.js. I already use those at work so I'd want to expose myself to different technologies.
3. The site should remain staticly generated but it should support the option of bundling Javascript for particular pages.
4. It must be cacheable (so hashed file names, year-long cache durations, etc).

So the work begins on rebuilding this website. Hopefully this is the last post on this version of the site. Hopefully, the next time I write something, the whole edit/release cycle will be different.

Now I just have to remember how to build and deploy this post 🤦
