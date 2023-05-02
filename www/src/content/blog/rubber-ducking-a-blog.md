---
title: "Rubber ducking a blog"
description: "Lorem ipsum dolor sit amet"
pubDate: "2017-04-21"
---

Hi, I'm Alex and I decided to start my own little corner on the Internet.

_&ldquo;But Alex, why start a blog?&rdquo;_ you might ask. And that's a good question.
I decided I wanted to start a personal site thinking that it's going to help me be a better engineer. Also, recently I re-read some older articles where [people]&nbsp;[asked] other people on the Internet to write more. So this is my take on it.

I'll start this tech blog with a post about how this site is built. Rubber ducking my blog if you will! It's a simple static site built using [Mustache], [Markdown] and tied together with some bash scripts.

_&ldquo;But Alex, why start everything from scratch when there are so many tools available to create static websites?&rdquo;_. True, there are a lot of static-site generators to choose from nowadays. But this post is about the journey (building the site) just as much as it is about the end result (you reading this webpage). I'll probably end up reinventing some of the functionality that a static-site generator would provide for free. And that's fine, at least I'll know how its insides work. 😄

## Requirements

Let's start by setting some functional requirements for the final site. In order, from "most important" to "would be cool":

1. It should be frictionless to start writing a new post.
2. Building it should be fast.
3. It must have an Index page that lists all the posts.
4. The post URLs should be clear of the `.html` extension. (why? I just like clean URLs)

Requirement #1 is the most important. If we don't get that ticked then my blog won't probably get a second post. The choice of tools helps me in this regard: Markdown (for the "words") and Mustache (for the chrome around the "words") are both very simple languages.

## Dev setup

I should have probably told you by now that this project is based on [Node.js] as it makes it easy to install dependencies for Markdown & Mustache. For those of you alergic to Node: it should be straight-forward to transpile this code to your chosen programming lanuage.

First off let's create a directory in which to keep all the files and initialize npm init it

```shell
$ mkdir the-blog
$ cd the-blog
$ npm init -y
```

Mustache and Markdown are now just an `npm install` away:

```shell
$ npm install --save-dev mustache marked
```

That's it for the dev setup. Now `mustache` and `marked` are available in the local `node_modules/` directory for our scripting pleasure.

## Rendering a blog post

{{=<% %>=}}
Let's start by writing a sample post and a mustache template to render it.

```shell
$ cat - > post.md <<EOF
# title
foo bar
EOF

$ cat - > post.mustache <<EOF
<!doctype html>
<html>
<body>
  {{> body }}
</body>
</html>
EOF
```

The weird `{{> body }}` syntax is how mustache renders partials. That bit of text is telling the renderer to take whatever the `body` variable is holding and treat it as a mustache template.
<%={{ }}=%>

Now that we have some sample files to play with, we need a Node script to process them into an HTML document.

```js
const marked = require("marked");
const mustache = require("mustache");
const fs = require("fs");

const contentFile = process.argv[2];
const templateFile = process.argv[3];

const content = fs.readFileSync(contentFile, "utf-8");
const template = fs.readFileSync(templateFile, "utf-8");

const body = marked(content);

const out = mustache.render(template, null, { body });

console.log(out);
```

The script is pretty simple. It takes two arguments: a markdown file and a mustache file (lines 5-9) and renders the mustache template, while passing the markdown content as the `body` fragment to the mustache renderer (lines 10-12). The second parameter on the `mustache.render` call is the viewmodel, but since we're only rendering static content we can just pass null. Finally, it prints the result to stdout.

Let's run it with the sample files:

```shell
$ node render.js post.md post.mustache
<!doctype html>
<html>
<body>
<h1 id="first-post">First post</h1>
<p>foo bar</p>
</body>
</html>
```

Cool, it works! As you can see, `marked` automatically assigned an `id` to the heading element. That's so we can link directly to it.

One problem with the script is that it can only process one pair of files at a time. We need to make it so that it processes all the posts at once. Let's use bash for that:

```shell
$ ls *.md | xargs -I file node render.js file post.mustache
<!doctype html>
<html>
<body>
<h1 id="title">title</h1>
<p>foo bar</p>
</body>
</html>
```

Perfect! The command renders all `.md` files in the current directory with the `post.mustache` template. It's a bit limiting as we're using a single template but it works for now. I'll save it as `build.sh` 🤓

```shell
$ echo "ls *.md | xargs -I file node render.js file post.mustache" > build.sh
$ chmod +x build.sh
```

## Cleaning up the URLs

In order to remove the `.html` extensions from the final URLs we can make use of the fact that webservers will serve an index file when a client requests a directory. This means that all our posts will actually be single `index.html` files in their own directories.

Let's move the `post.md` to its own directory:

```shell
$ mkdir -p blog/hello-world
$ mv post.md blog/hello-world
```

While we're at it let's also move the render command into a separate bash script to make it easier to modify:

```shell
$ cat - > render.sh <<EOF
node render.js $1 post.mustache
EOF
$ chmod +x render.sh
```

`$1` is the argument number for the filename that `xargs` will pass to our script. This makes `build.sh` a lot simpler:

```bash
find blog -name "*.md" -print0 | xargs -0 -n 1 ./render.sh
```

I've switched the `ls` command to `find` in order to search for markdown files in the tree rooted at `blog/`. Also `find` is separating the files with `NULL` characters instead of whitespace so that we can properly handle filenames with spaces in them. `xargs` changed a bit as well: it uses `NULL` as its delimiter (so it matches up with `find`) and also only passes one filename per run of `./render.sh` (otherwise `xargs` would try to pass multiple files to the script and we aren't handling that case well... or at all).

What's left is to get `render.sh` to actually print to a file instead of stdout:

```bash
TITLE=$(dirname "$1")
OUT_DIR="public/$TILTE"
mkdir -p "$OUT_DIR"
node render.js "$1" post.mustache > "$OUT_DIR/index.html"
```

Simple. The script will mirror the tree under `blog` to the `public` directory. For example, if `$1` is `blog/hello-world/post.md`, then `render.sh` creates `public/blog/hello-world/index.html`. Notice the quotes around filenames. We need them in order to be able to handle directories that have spaces.

## Creating a blog index

Now that we can actually write posts and have them rendered as HTML documents in the `public/blog/` directory, it's time we create an index page that lists all the posts.

Thankfully we get this for free: if we create a markdown file in the `blog/` directory then the `render.sh` script is going to render it to `public/blog/index.html`. Let's test it out:

```shell
$ cat - > blog/index.md <<EOF
# The blog index!
- [Hello world](hello-world)
EOF

$ sh build.sh
$ tree public/
public/
└── blog
    ├── hello-world
    │   └── index.html
    └── index.html

2 directories, 2 files
$ cat public/blog/index.html
<!doctype html>
<html>
<body>
<h1 id="the-blog-index-">The blog index!</h1>
<ul>
<li><a href="hello-world">Hello world</a></li>
</ul>
</body>
</html>
```

Cool, that worked! But, there's one problem. If there are multiple files in the `blog/` folder then they will compete for `public/blog/index.html` and overwrite each other. Actually this problem will occur if any folder under `blog` contains more than one markdown file. 😱 A quick fix would be the prevent file overwrite (aka "clobbering") in bash, but that means only one markdown file per post. Well... I can live with that, let's update `render.sh`:

```bash
set -o noclobber

TITLE=$(dirname "$1")
OUT_DIR="public/$TITLE"
mkdir -p "$OUT_DIR"
node render.js "$1" post.mustache > "$OUT_DIR/index.html"
```

The only downside to this approach is that the index file needs to be updated manually with links to new posts.

## Conclusion & going forward

That's it for now. It's a very basic system, but it ticks all the requirements for me. Though, there are a few improvements that I want to bring to it in the — hopefully near — future:

- auto-generate the blog index
- record creation/update date of posts and render them in the template (maybe get this data from git?)
- set page title to the current post
- have different page templates for the index and posts files
- have a table of content at the start of a blog post
- tags for posts

For now, I'm going to add a bit of CSS to the pages and work on the base mustache template to make it look more like a full HTML document.

Thanks for reading! You can have a look at the complete source code of this blog on my GitLab https://gitlab.com/alexghr/alexghr.gitlab.io

[people]: http://www.drmaciver.com/2014/07/you-should-write-more/
[asked]: http://jsomers.net/blog/more-people-should-write
[Node.js]: https://nodejs.org
[Markdown]: http://commonmark.org
[Mustache]: https://mustache.github.io
