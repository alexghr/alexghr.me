module.exports = {
  pageTitle: "Rubber ducking a blog",
  url: "/blog/rubber-ducking-a-blog",
  extends: __dirname + '/../../../_templates/article',
  date: '2017-04-21',
  list: true,
  partials: {
    article: __dirname + '/post.md'
  }
}
