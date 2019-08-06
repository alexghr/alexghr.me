const fs = require('fs')

module.exports = {
  pageTitle: "Rubber ducking a blog. Part 2 GitLab Pages",
  url: "/blog/rubber-ducking-a-blog-2",
  extends: __dirname + '/../../../_templates/article',
  date: '2017-05-04',
  list: false,
  partials: {
    article: __dirname + '/post.md'
  },
  assets: [__dirname + '/custom-domain.png']
}
